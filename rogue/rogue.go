package rogue

import (
    "bytes"
    "github.com/go-chat-bot/bot"
    "regexp"
    "strconv"
    "time"
    "math/rand"
//  "github.com/go-chat-bot/bot/irc"
//  "fmt"
//  "net/http"
//  "net/url"
//  "io/ioutil"
//  "strings"
//  "os"
//  "encoding/json"
)

type gameStateId int

const (
    // Game state enumeration
    IdleState gameStateId = iota
    EncounterState
)

type class int

const (
    // Game state enumeration
    Fighter class = iota
    Rogue
)

const (
    // Regex patterns
    dicePattern = "([1-9]\\d*)?d([1-9]\\d*)(([+-])([1-9]\\d*)d([1-9]\\d*))?"
)

var (
    // Random number generator
    rnd = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
    // This holds the current game state. 
    currentGameStateId gameStateId = IdleState
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
}

func (char character) classString() string {
    var buffer bytes.Buffer

    if char.Class == Fighter {
        buffer.WriteString("Fighter")
    } else if char.Class == Rogue {
        buffer.WriteString("Rogue")
    }

    return buffer.String()
}

type idleState struct {
}

func (is idleState) describe() string {
    return "You are in the town of Waterdeep taking care of various kinds of \"business\"...  If you would to join a party going to the nearby dungeon, type !joinparty"
}

func (is idleState) name() string {
    return "Idle"
}

type encounterState struct {
}

func (en encounterState) describe() string {
    return "Encounter State"
}

func (en encounterState) name() string {
    return "Encounter"
}

type characterCreationState struct {
}

func (cc characterCreationState) describe() string {
    return "Character Creation State"
}

func (cc characterCreationState) name() string {
    return "Character Creation"
}

func rollCharacter(command *bot.Cmd) (msg string) {
    var buffer bytes.Buffer
    // Set who is creating a character.
    currentUserCreatingCharacter = command.User
    // Roll 3d6 for Str and Dex.
    str := rollDice("3d6")
    dex := rollDice("3d6")
    // Figure out what class.
    var selectedClass class
    if len(command.Args) == 0 || command.Args[0] == "Fighter" {
        selectedClass = Fighter
    } else if command.Args[0] == "Rogue" {
        selectedClass = Rogue
    }
    // Add the character to the list of characters.
    characters[command.User.Nick] = character{
        command.User.Nick, selectedClass, str, dex,
    }
    // Build the message.
    buffer.WriteString(command.User.Nick)
    buffer.WriteString(" is now a <")
    buffer.WriteString(characters[command.User.Nick].classString())
    buffer.WriteString("> with STR <")
    buffer.WriteString(strconv.Itoa(str))
    buffer.WriteString("> and DEX <")
    buffer.WriteString(strconv.Itoa(dex))
    buffer.WriteString(">")

    return buffer.String()
}

func createCharacter(command *bot.Cmd) (msg string, err error) {
    var buffer bytes.Buffer
    // To create a character we need to be in an idle state.  
    if currentGameStateId != IdleState {
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
        buffer.WriteString(">")
    } else {
        buffer.WriteString("No character exists for you. Try !createcharacter")
    }

    return buffer.String(), nil
}

func joinParty(command *bot.Cmd) (msg string, err error) {
    var buffer bytes.Buffer
    if currentGameStateId != IdleState {
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

func leaveParty(command *bot.Cmd) (msg string, err error) {
    var buffer bytes.Buffer
    if currentGameStateId != IdleState || !isInParty(command) {
        buffer.WriteString(command.User.Nick)
        buffer.WriteString(" can't leave the party now")
    } else if _, ok := characters[command.User.Nick]; ok {
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
                buffer.WriteString(" ,")
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

// Initializes or Sets up the game 
func initializeGame() {
    // Initialize idle state 
    is := idleState{}
    gameStates = append(gameStates, is)

    // Initialize encounter state 
    en := encounterState{}
    gameStates = append(gameStates, en)

    // Initialize character creation state 
    cc := characterCreationState{}
    gameStates = append(gameStates, cc)
}

func init() {

    initializeGame()

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
}
