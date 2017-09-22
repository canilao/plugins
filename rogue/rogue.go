package rogue

import (
    "bytes"
    "github.com/go-chat-bot/bot"
    "regexp"
    "strconv"
    "time"
    "math/rand"
    "strings"
//  "github.com/go-chat-bot/bot/irc"
//  "fmt"
//  "net/http"
//  "net/url"
//  "io/ioutil"
//  "os"
//  "encoding/json"
)

type gameStateId int

const (
    // Game state enumeration
    TownState gameStateId = iota
    TravelState
)

type direction int

const (
    // Direction enumeration 
    North direction = iota
    East
    South
    West
)

type class int

const (
    // Game state enumeration
    Fighter class = iota
    Rogue
    Cleric
)

func convertClassEnumToString(classEnum class) string {
    var buffer bytes.Buffer

    if classEnum == Fighter {
        buffer.WriteString("Fighter")
    } else if classEnum == Rogue {
        buffer.WriteString("Rogue")
    } else if classEnum == Cleric {
        buffer.WriteString("Cleric")
    }

    return buffer.String()
}

func convertClassStringToEnum(classString string) class {
    var selectedClass class = Fighter
    if strings.ToLower(classString) == "fighter" {
        selectedClass = Fighter
    } else if strings.ToLower(classString) == "rogue" {
        selectedClass = Rogue
    } else if strings.ToLower(classString) == "cleric" {
        selectedClass = Cleric
    }
    return selectedClass
}

const (
    // Regex patterns
    dicePattern = "([1-9]\\d*)?d([1-9]\\d*)(([+-])([1-9]\\d*)d([1-9]\\d*))?"
)

var (
    // Random number generator
    rnd = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
    // This holds the current game state. 
    currentGameStateId gameStateId = TownState
    // This array holds the state objects.
    gameStates []gameState
    // Current user creating a character.
    currentUserCreatingCharacter *bot.User
    // This map holds the created characters 
    characters map[string]character = make(map[string]character)
    // This array holds all of the party members. 
    theParty []character
)

type gameState interface {
    // The name of the state.
    name() string
    // This describes the current state.
    describe() string
    // This describes the transitioning into the state. 
    describeTransition() string
    // Handles the movement 
    move(dir direction, command *bot.Cmd) string
}

type character struct {
    // The character's name (the user's nick)
    Name string
    // The character class
    Class class
    // The character's brawn
    Strength int
    // The character's agility
    Dexterity int
    // The character's "knowing of all things"
    Wisdom int
}

func (char character) classString() string {
    return convertClassEnumToString(char.Class)
}

type townState struct {
}

func (ts townState) describeTransition() string {
    return "After a long trek through the nearby canyon, you finally enter the town of Waterdeep. The town is bustling with activity. Here is your chance to take care of some business."
}


func (ts townState) describe() string {
    return "You are in the town of Waterdeep taking care of various kinds of \"business\"...  <joinparty> to visit the nearby dungeon to the <north>"
}

func (ts townState) name() string {
    return "TownState"
}

func (ts townState) move(dir direction, command *bot.Cmd) string {
    var buffer bytes.Buffer
    if dir == North {
        buffer.WriteString(changeState(TravelState))
    } else if dir == East {
        buffer.WriteString("Invalid direction")
    } else if dir == South {
        buffer.WriteString("Invalid direction")
    } else if dir == West {
        buffer.WriteString("Invalid direction")
    }
    return buffer.String()
}

type travelState struct {
}

func (tv travelState) name() string {
    return "Leave Town"
}

func (tv travelState) describeTransition() string {
    var buffer bytes.Buffer
    buffer.WriteString(theParty[0].Name)
    buffer.WriteString(" leads the party out of town towards the Black Water Mines")
    return buffer.String()
}

func (tv travelState) describe() string {
    return "You've hiked several hours through the canyon that began just outside of town. Now, just before you is a rather small entrance to the Black Water Mines. As you look closer, you can smell a musty, dank odor as the wind howls, blowing outward from the entrance. What could be lurking inside? !move <north> into the mines or !move <south> to go back to town."
}

func (tv travelState) move(dir direction, command *bot.Cmd) string {
    var buffer bytes.Buffer
    if dir == North {
        buffer.WriteString("Invalid direction")
    } else if dir == East {
        buffer.WriteString("Invalid direction")
    } else if dir == South {
        buffer.WriteString(changeState(TownState))
    } else if dir == West {
        buffer.WriteString("Invalid direction")
    }
    return buffer.String()
}

func rollCharacter(command *bot.Cmd) (msg string) {
    var buffer bytes.Buffer
    // Set who is creating a character.
    currentUserCreatingCharacter = command.User
    // Roll 3d6 for Str and Dex.
    str := rollDice("3d6")
    dex := rollDice("3d6")
    wis := rollDice("3d6")
    // Figure out what class.
    var selectedClass class = Fighter
    if len(command.Args) > 0 {
        selectedClass = convertClassStringToEnum(command.Args[0])
    }
    // Add the character to the list of characters.
    characters[command.User.Nick] = character{
        command.User.Nick, selectedClass, str, dex, wis,
    }
    // Build the message.
    buffer.WriteString(command.User.Nick)
    buffer.WriteString(" is now a <")
    buffer.WriteString(characters[command.User.Nick].classString())
    buffer.WriteString("> with STR <")
    buffer.WriteString(strconv.Itoa(str))
    buffer.WriteString(">, DEX <")
    buffer.WriteString(strconv.Itoa(dex))
    buffer.WriteString(">, and WIS <")
    buffer.WriteString(strconv.Itoa(wis))
    buffer.WriteString(">")

    return buffer.String()
}

func createCharacter(command *bot.Cmd) (msg string, err error) {
    var buffer bytes.Buffer
    // To create a character we need to be in an idle state.  
    if currentGameStateId != TownState || isInParty(command) {
        buffer.WriteString(command.User.Nick)
        buffer.WriteString(" cannot create a character now")
    } else {
        buffer.WriteString(rollCharacter(command))
    }

    return buffer.String(), nil
}

func describe(command *bot.Cmd) (msg string, err error) {
    return gameStates[currentGameStateId].describe(), nil
}

func characterStats(command *bot.Cmd) (msg string, err error) {
    var buffer bytes.Buffer
    if _, ok := characters[command.User.Nick]; ok {
        buffer.WriteString(characters[command.User.Nick].classString())
        buffer.WriteString(": STR <")
        buffer.WriteString(strconv.Itoa(characters[command.User.Nick].Strength))
        buffer.WriteString("> DEX <")
        buffer.WriteString(strconv.Itoa(characters[command.User.Nick].Dexterity))
        buffer.WriteString("> WIS <")
        buffer.WriteString(strconv.Itoa(characters[command.User.Nick].Wisdom))
        buffer.WriteString(">")
    } else {
        buffer.WriteString("No character exists for you. Try !createcharacter")
    }

    return buffer.String(), nil
}

func joinParty(command *bot.Cmd) (msg string, err error) {
    var buffer bytes.Buffer
    if currentGameStateId != TownState {
        buffer.WriteString(command.User.Nick)
        buffer.WriteString(" can't join the party now")
    } else if _, ok := characters[command.User.Nick]; ok {
        // Add the character to the party 
        theParty = append(theParty, characters[command.User.Nick])
        if len(theParty) == 0 {
            // Create the message.
            buffer.WriteString(command.User.Nick)
            buffer.WriteString(" is party leader")
        } else {
            // Create the message.
            buffer.WriteString(command.User.Nick)
            buffer.WriteString(" joined the party")
        }
    } else {
        buffer.WriteString("No character exists for you. Try !createcharacter")
    }

    return buffer.String(), nil
}

func isInParty(command *bot.Cmd) (bool) {
    inParty := false
    for _, element := range theParty {
        inParty = inParty || (element.Name == command.User.Nick)
    }
    return inParty
}

func removeFromParty(command *bot.Cmd) {
    var index int
    for i, v := range theParty {
        if v.Name == command.User.Nick {
            index = i
            break
        }
    }
    theParty = append(theParty[:index], theParty[index+1:]...)
}

func leaveParty(command *bot.Cmd) (msg string, err error) {
    var buffer bytes.Buffer
    if currentGameStateId != TownState || !isInParty(command) {
        buffer.WriteString(command.User.Nick)
        buffer.WriteString(" can't leave the party now")
    } else if _, ok := characters[command.User.Nick]; ok {
        // Physically removes the user's character from the party array.
        removeFromParty(command)
        buffer.WriteString(command.User.Nick)
        buffer.WriteString(" left the party")
    } else {
        buffer.WriteString("Your character does not exist.")
    }

    return buffer.String(), nil
}

func listParty(command *bot.Cmd) (msg string, err error) {
    var buffer bytes.Buffer
    first := true
    if len(theParty) == 0 {
        buffer.WriteString("No one is in the party")
    } else {
        for _, element := range theParty {
            if !first {
                buffer.WriteString(", ")
            }
            buffer.WriteString(element.Name)
            buffer.WriteString(" <")
            buffer.WriteString(element.classString())
            buffer.WriteString(">")
            first = false
        }
    }

    return buffer.String(), nil
}

func move(command *bot.Cmd) (msg string, err error) {
    var buffer bytes.Buffer
    if len(theParty) == 0 {
        buffer.WriteString("No one is in the party")
    } else if theParty[0].Name != command.User.Nick {
        buffer.WriteString("You are not the leader of the party")
    } else {
        if len(command.Args) == 0  {
            buffer.WriteString("Invalid direction")
        } else if command.Args[0] == "north" {
            buffer.WriteString(gameStates[currentGameStateId].move(North, command))
        } else if command.Args[0] == "east" {
            buffer.WriteString(gameStates[currentGameStateId].move(East, command))
        } else if command.Args[0] == "south" {
            buffer.WriteString(gameStates[currentGameStateId].move(South, command))
        } else if command.Args[0] == "west" {
            buffer.WriteString(gameStates[currentGameStateId].move(West, command))
        } else {
            buffer.WriteString("Invalid direction")
        }
    }
    return buffer.String(), nil
}

func shop(command *bot.Cmd) (msg string, err error) {
    return "A bunch of crap!", nil
}

func dickSlap(command *bot.Cmd) (msg string, err error) {
    return "Boticus slowly whips it out, then he slaps syntac, gently. Just the way he likes.", nil 
}

func rollDice(dice string) (int) {
    re := regexp.MustCompile(dicePattern)
    groups := re.FindStringSubmatch(dice)
    total := 0
    count1, _ := strconv.Atoi(groups[1])
    dice1, _ := strconv.Atoi(groups[2])
    accumulator1 := 0

    for j := 0; j < count1; j++ {
        roll := rnd.Intn(dice1)
        roll++
        accumulator1 += roll
    }

    if len(groups) >= 6 {
        count2, _ := strconv.Atoi(groups[5])
        dice2, _ := strconv.Atoi(groups[6])
        accumulator2 := 0

        for j := 0; j < count2; j++ {
            roll := rnd.Intn(dice2)
            roll++
            accumulator2 += roll
        }

        if groups[3] == "+" {
            total = accumulator1 + accumulator2
        } else {
            total = accumulator1 - accumulator2
            if total <= 0 { total = 0 }
        }
    } else {
        total = accumulator1
    }

    return total
}

func changeState(newStateId gameStateId) string {
    currentGameStateId = newStateId
    return gameStates[currentGameStateId].describeTransition()
}

// Initializes or Sets up the game 
func initializeGame() {
    // Initialize town state 
    ts := townState{}
    gameStates = append(gameStates, ts)

    // Initialize travel state 
    tv := travelState{}
    gameStates = append(gameStates, tv)
}

func init() {

    initializeGame()

    bot.RegisterCommand(
        "dickslap",
        "Give's syntac what he wants",
        "",
        dickSlap)

    bot.RegisterCommand(
        "describe",
        "Describes the party's current location",
        "",
        describe)

    bot.RegisterCommand(
        "createcharacter",
        "Creates a new character",
        "",
        createCharacter)

    bot.RegisterCommand(
        "characterstats",
        "Print character's stats",
        "",
        characterStats)

    bot.RegisterCommand(
        "joinparty",
        "Join the party",
        "",
        joinParty)

    bot.RegisterCommand(
        "leaveparty",
        "Leave the party",
        "",
        leaveParty)

    bot.RegisterCommand(
        "listparty",
        "Lists the members of the party",
        "",
        listParty)

    bot.RegisterCommand(
        "move",
        "Moves the party in the given direction",
        "",
        move)

    bot.RegisterCommand(
        "shop",
        "Lists items in the Waterdeep marketplace. You can purchase <weapons>, <armor>",
        "",
        shop)
}
