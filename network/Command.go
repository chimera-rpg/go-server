package network

import (
	"time"

	"github.com/chimera-rpg/go-server/data"
)

// Command is our interface for all commands.
type Command interface {
	GetType() uint32
}

// CommandBasic represent very simple transmissions between the server and
// the client. This is used for disconnects among other things.
type CommandBasic struct {
	Type   uint8
	String string
}

// GetType returns TypeBasic
func (c CommandBasic) GetType() uint32 {
	return TypeBasic
}

// CommandHandshake represents the handshake between the server and the client
// so as to ensure compatibility.
type CommandHandshake struct {
	Version int
	Program string
}

// GetType returns TypeHandshake
func (c CommandHandshake) GetType() uint32 {
	return TypeHandshake
}

// Versioning. This should probably be different.
const (
	Version = iota
)

// CommandFeatures handles the communication of the features of the server, such as animations sizes, to the client.
type CommandFeatures struct {
	AnimationsConfig data.AnimationsConfig
	TypeHints        map[uint32]string
	Slots            map[uint32]string
}

// GetType returns TypeFeatures
func (c CommandFeatures) GetType() uint32 {
	return TypeFeatures
}

// CommandViewport requests a specific viewport size from the server. This mainly intended for small displays that can't see as much data as they can receive.
type CommandViewport struct {
	Height, Width, Depth uint8
}

// GetType returns TypeViewport
func (c CommandViewport) GetType() uint32 {
	return TypeViewport
}

// CommandLogin handles the process of logging in, registering, recovering
// a password via email, and even deleting the account.
type CommandLogin struct {
	Type  uint8
	User  string
	Pass  string
	Email string
}

// GetType returns TYPE_LOGIN
func (c CommandLogin) GetType() uint32 {
	return TypeLogin
}

// These are the CommandLogin Types
const (
	Query = iota
	Login
	Register
	Delete
)

// CommandRejoin signifies the client is rejoining a loaded character.
type CommandRejoin struct {
}

// GetType returns TypeRejoin
func (c CommandRejoin) GetType() uint32 {
	return TypeRejoin
}

// CommandQueryCharacters is sent by the client to ask for their characters.
type CommandQueryCharacters struct {
}

// GetType returns TypeCharacter
func (c CommandQueryCharacters) GetType() uint32 {
	return TypeQueryCharacters
}

// CommandQueryGenera is sent by the client to request genera. The server responds with the same message type populated with the genera.
type CommandQueryGenera struct {
	Genera []Genus
}

// GetType returns TypeQueryGenera
func (c CommandQueryGenera) GetType() uint32 {
	return TypeQueryGenera
}

// CommandQuerySpecies is sent by the client to request the species for a given genera. The server responds with the same message type populated with the species.
type CommandQuerySpecies struct {
	Genus   string
	Species []Species
}

// GetType returns TypeQuerySpecies
func (c CommandQuerySpecies) GetType() uint32 {
	return TypeQuerySpecies
}

// CommandQueryCulture works like Species, wow.
type CommandQueryCulture struct {
	Genus   string
	Species string
	// Cultures []Culture
}

// GetType returns TypeQueryCulture
func (c CommandQueryCulture) GetType() uint32 {
	return TypeQueryCulture
}

// CommandQueryTraining works like Culture.
type CommandQueryTraining struct {
	Genus   string
	Species string
	Culture string
	// Trainings []Training
}

// GetType returns TypeQueryTraining
func (c CommandQueryTraining) GetType() uint32 {
	return TypeQueryTraining
}

// CommandCharacter is sent by the server to show the characters the player has. If sent by the client with the Delete bool true, the character will be deleted.
type CommandCharacter struct {
	Name        string
	Attributes  data.AttributeSets
	AnimationID uint32
	FaceID      uint32
	Delete      bool
}

// GetType returns TypeCharacter
func (c CommandCharacter) GetType() uint32 {
	return TypeCharacter
}

// CommandCreateCharacter creates a given character.
type CommandCreateCharacter struct {
	Name     string
	Genus    string
	Species  string
	Culture  string
	Training string
}

// GetType returns TypeCreateCharacter
func (c CommandCreateCharacter) GetType() uint32 {
	return TypeCreateCharacter
}

// Genus represents a single genus in genera.
type Genus struct {
	Name        string
	Description string
	Attributes  data.AttributeSets
	AnimationID uint32
	FaceID      uint32
}

// Species is a species, wow.
type Species struct {
	Name        string
	Description string
	Attributes  data.AttributeSets
	AnimationID uint32
	FaceID      uint32
}

// CommandSelectCharacter is sent to the server to select a character for play. It is sent by the server to indicate the selection is valid and the client should transition to game state.
type CommandSelectCharacter struct {
	Name string
}

func (c CommandSelectCharacter) GetType() uint32 {
	return TypeSelectCharacter
}

// Our basic return types
const (
	Nokay = iota
	Okay
	OnMap
	Set
	Get
	Reject
	Cya
)

// CommandAnimation is for setting and/or getting animation ID->FaceIDs->Frames
type CommandAnimation struct {
	Type        uint8                       // ONMAP->, SET->, ->GET
	AnimationID uint32                      // Animation ID in question
	Faces       map[uint32][]AnimationFrame // FaceID to Frames
	RandomFrame bool                        // Whether to start the animation at a random frame.
}

// AnimationFrame represents an imageID and how long it should play.
type AnimationFrame struct {
	ImageID uint32
	Time    int
	Y, X    int8 // Allow X a Y adjustments from -128 to 128
}

// GetType returns TypeAnimation.
func (c CommandAnimation) GetType() uint32 {
	return TypeAnimation
}

// Our Graphics data types.
const (
	GraphicsPng = iota
)

// CommandGraphics are for setting and requesting images.
type CommandGraphics struct {
	Type       uint8  // SET->, ->GET
	GraphicsID uint32 //
	DataType   uint8  // GRAPHICS_PNG, ...
	Data       []byte
}

// GetType returns TypeGraphics.
func (c CommandGraphics) GetType() uint32 {
	return TypeGraphics
}

// CommandAudio is for setting and/or getting audio ID->SoundIDs->Sounds
type CommandAudio struct {
	Type    uint8
	AudioID uint32
	Sounds  map[uint32][]AudioSound
}

// GetType returns TypeAudio
func (c CommandAudio) GetType() uint32 {
	return TypeAudio
}

// AudioSound contains a soundID and its text representation.
type AudioSound struct {
	SoundID uint32
	Text    string
}

// Our Audio data types.
const (
	SoundOgg = iota
	SoundFlac
)

// CommandSound is for setting and requesting sound files.
type CommandSound struct {
	Type     uint8
	SoundID  uint32
	DataType uint8 // SoundOgg, ...
	Data     []byte
}

// GetType returns TypeAudio.
func (c CommandSound) GetType() uint32 {
	return TypeSound
}

// Our CommandMap.Type constants.
const (
	Travel = iota
)

// CommandMap is a basic command for creating a map of a given name and ID at provided dimensions.
type CommandMap struct {
	Type                                  uint8 // TRAVEL
	MapID                                 uint32
	Name                                  string // target map name
	Height                                int
	Width                                 int
	Depth                                 int
	Outdoor                               bool
	OutdoorRed, OutdoorGreen, OutdoorBlue uint8
	AmbientRed, AmbientGreen, AmbientBlue uint8
}

// GetType returns TypeMap
func (c CommandMap) GetType() uint32 {
	return TypeMap
}

// CommandTiles is a batch update of all tile updates.
type CommandTiles struct {
	TileUpdates  []CommandTile
	LightUpdates []CommandTileLight
	SkyUpdates   []CommandTileSky
}

// GetType returns TypeTiles
func (c CommandTiles) GetType() uint32 {
	return TypeTiles
}

// CommandTile is a list of tiles at a given Tile. This might be expanded to also have a brightness/visibility value.
type CommandTile struct {
	X, Y, Z   uint32
	ObjectIDs []uint32
}

// GetType returns TypeTileUpdate
func (c CommandTile) GetType() uint32 {
	return TypeTileUpdate
}

// CommandTileLight is the brightness and color value of a given tile.
type CommandTileLight struct {
	X, Y, Z uint32
	R, G, B uint8
}

// GetType returns TypeTileLight
func (c CommandTileLight) GetType() uint32 {
	return TypeTileLight
}

// CommandTileSky is the sky value of a given tile.
type CommandTileSky struct {
	X, Y, Z uint32
	Sky     float64
}

// GetType returns TypeTileSky
func (c CommandTileSky) GetType() uint32 {
	return TypeTileSky
}

// CommandObject is the command type used to create, delete, and update objects.
type CommandObject struct {
	ObjectID uint32 // id of target object
	Payload  CommandObjectPayload
}

// GetType returns TypeObjectUpdate
func (c CommandObject) GetType() uint32 {
	return TypeObjectUpdate
}

// CommandObjectPayload is a generic interface for actual payloads.
type CommandObjectPayload interface {
}

// CommandObjectPayloadCreate is the type for creating a new object.
type CommandObjectPayloadCreate struct {
	TypeID               uint8
	AnimationID          uint32
	FaceID               uint32
	Height, Width, Depth uint8
	Reach                uint8 // Reach is really only used by the player's object.
	Opaque               bool
}

// CommandObjectPayloadDelete is the type indicating that an object should be deleted.
type CommandObjectPayloadDelete struct {
}

// CommandObjectPayloadAnimate is the type used for updating an object's animation and face.
type CommandObjectPayloadAnimate struct {
	AnimationID uint32 //
	FaceID      uint32 //
}

// CommandObjectPayloadInfo is the type used for updating an object's information.
type CommandObjectPayloadInfo struct {
	Info []data.ObjectInfo
}

// CommandObjectPayloadViewTarget is the type used for marking a given object as the client's view target. It additionally sends the view range and reach of the given object.
type CommandObjectPayloadViewTarget struct {
	Height, Width, Depth uint8
}

// Our Object types (unused)
const (
	ObjectCreate     = iota // used to create an object with given id.
	ObjectDelete            // used to completely delete given object.
	ObjectAnimate           // whether used to set AnimationID and FaceID.
	ObjectViewTarget        // used to target the object as the client's view.
)

// CommandInspect is used to request an inspect of an object. This will cause a CommandObject with an info payload to be sent if valid.
type CommandInspect struct {
	ObjectID uint32
}

// GetType returns TypeInspect
func (c CommandInspect) GetType() uint32 {
	return TypeInspect
}

// CommandCmd is used for player commands to interact with the game world.
type CommandCmd struct {
	Cmd  int
	Data interface{}
}

// GetType returns TypeCmd
func (c CommandCmd) GetType() uint32 {
	return TypeCmd
}

// Our various CommandCmd.Cmd values
const (
	North = iota
	South
	East
	West
	Northeast
	Northwest
	Southeast
	Southwest
	Up
	Down
	Brace
	Drop
	Attack
	Quit
	Wizard
)

// CommandClearCmd is used to clear the enter command queue.
type CommandClearCmd struct{}

// GetType returns TypeClearCmd
func (c CommandClearCmd) GetType() uint32 {
	return TypeClearCmd
}

// CommandExtCmd is used for extended player commands with variadic inputs.
type CommandExtCmd struct {
	Cmd  string
	Args []string
}

// GetType returns TypeExtCmd
func (c CommandExtCmd) GetType() uint32 {
	return TypeExtCmd
}

// CommandRepeatCmd is used to send repeating versions of the above CommandCmds.
type CommandRepeatCmd struct {
	Cmd    int
	Cancel bool // If the action should be canceled (used for canceling the repeat)
	Data   interface{}
}

// GetType returns TypeRepeatCmd
func (c CommandRepeatCmd) GetType() uint32 {
	return TypeRepeatCmd
}

// Our CommandMessage.Type values
const (
	ServerMessage = iota
	MapMessage
	TargetMessage // Message for the client's target (their character).
	PCMessage
	NPCMessage
	PartyMessage
	GuildMessage
	ChatMessage
	LocalMessage // Client-side only messaging
)

// CommandMessage is used for most forms of messaging.
type CommandMessage struct {
	Type         int    // Message type, representing the context of the message.
	From         string // From, if it matters
	FromObjectID uint32 // For NPC messaging.
	Title        string // Title of the message. Optional.
	Body         string // Body of the message.
}

// GetType returns TypeMessage
func (c CommandMessage) GetType() uint32 {
	return TypeMessage
}

// Our CommandNoise values
const (
	GenericNoise = iota
	MapNoise
	ObjectNoise
)

// CommandNoise is used for playing sounds and showing the on-screen representation of them.
type CommandNoise struct {
	Type     int
	AudioID  uint32  // The audio ID to be played.
	SoundID  uint32  // The specific sound ID that should be played.
	ObjectID uint32  // ObjectID, for sounds eminating from an object.
	X, Y, Z  uint32  // The origin of the sound, if not part of an ObjectID.
	Volume   float32 // The volume, from 0 to 1.
}

// GetType returns TypeNoise
func (c CommandNoise) GetType() uint32 {
	return TypeNoise
}

// CommandMusic is used for playing music.
type CommandMusic struct {
	Type     int
	AudioID  uint32  // The audio ID to be played.
	SoundID  uint32  // The specific sound ID that should be played.
	ObjectID uint32  // ObjectID, for music eminating from an object.
	X, Y, Z  uint32  // The origin of the music.
	Volume   float32 // The volume, from 0 to 1.
	Loop     int8    // The loop count, -1 being infinite.
	Stop     bool    // If the music should be stoppped.
}

// GetType returns TypeMusic
func (c CommandMusic) GetType() uint32 {
	return TypeMusic
}

// CommandStatus is used to notify the client of status effects as well as to let the server know we want to add/remove particular status effects.
type CommandStatus struct {
	Type   data.StatusType // StatusType.
	Active bool            // If it is (or desired to be) active or not.
}

// GetType returns TypeStatus
func (c CommandStatus) GetType() uint32 {
	return TypeStatus
}

// CommandStamina is used to notify the client of changes in its target's stamina.
type CommandStamina struct {
	Stamina    time.Duration
	MaxStamina time.Duration
}

// GetType returns TypeStamina
func (c CommandStamina) GetType() uint32 {
	return TypeStamina
}

// CommandAttack is used to send an attack character action.
type CommandAttack struct {
	Direction int    // Direction of the attack. Used with direction melee swings.
	Y, X, Z   uint32 // Specific Y, X, Z to target. Used with targeted range.
	Target    uint32 // Object ID to target.
}

// GetType returns TypeAttack
func (c CommandAttack) GetType() uint32 {
	return TypeAttack
}

// CommandDamage is used to send the results of damage.
type CommandDamage struct {
	Target          uint32 // Object ID to target.
	Type            data.AttackType
	StyleDamage     map[data.AttackStyle]float64
	AttributeDamage float64 // FIXME: This needs to be a map[AttributeId]float64
}

// GetType returns TypeDamage
func (c CommandDamage) GetType() uint32 {
	return TypeDamage
}

// CommandInteract
type CommandInteract struct {
	Target uint32 // Target Object ID to target.
	Type   int
}

// GetType returns TypeInteract
func (c CommandInteract) GetType() uint32 {
	return TypeInteract
}

const (
	InspectInteraction = iota
	PickupInteraction
	DropInteraction
	EquipInteraction
	UnequipInteraction
	ActivateInteraction
)

// A list of all our command types.
const (
	TypeBasic = iota
	TypeHandshake
	TypeFeatures
	TypeLogin
	TypeRejoin
	TypeQueryCharacters
	TypeQueryGenera
	TypeQuerySpecies
	TypeQueryCulture
	TypeQueryTraining
	TypeCharacter
	TypeCreateCharacter
	TypeSelectCharacter
	TypeData
	TypeTiles
	TypeTileUpdate
	TypeTileLight
	TypeTileSky
	TypeObjectUpdate
	TypeInventoryUpdate
	TypeInspect
	TypeStatus
	TypeMap
	TypeCmd
	TypeClearCmd
	TypeExtCmd
	TypeRepeatCmd
	TypeMessage
	TypeViewport
	TypeStamina
	TypeAttack
	TypeDamage
	TypeInteract
	// Graphics-related
	TypeGraphics
	TypeAnimation

	// Audio-related
	TypeAudio
	TypeSound
	TypeNoise
	TypeMusic
)
