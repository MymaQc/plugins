// Command csharp-gen emits the supported C# surface directly from Dragonfly's Go AST.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"image/color"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bedrock-gophers/plugins/internal/blockstate"
	_ "github.com/df-mc/dragonfly/server/block"
	dfitem "github.com/df-mc/dragonfly/server/item"
	_ "github.com/df-mc/dragonfly/server/item/enchantment"
	dfpotion "github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	_ "github.com/df-mc/dragonfly/server/world/biome"
	dfsound "github.com/df-mc/dragonfly/server/world/sound"
)

type method struct {
	Name         string
	Parameters   []parameter
	Subscription uint64
}

type parameter struct {
	Name string
	Type string
}

type commandInterface struct {
	Name       string
	Embeddings []string
	Methods    []commandMethod
}

type commandMethod struct {
	Name       string
	Parameters []parameter
	ReturnType string
}

type generatedFile struct {
	Path    string
	Content []byte
}

type cubeSpec struct {
	Faces []string
}

type encodedBlock struct {
	Name          string
	Identifier    string
	PropertiesNBT []byte
}

type blockField struct {
	Name string
	Type string
}

type blockFieldValue struct {
	Bool bool
	Int  int
}

type encodedBlockState struct {
	encodedBlock
	Values []blockFieldValue
}

type blockTypeSpec struct {
	Name   string
	Fields []blockField
	States []encodedBlockState
}

type blockDecodeCase struct {
	Identifier  string
	State       int
	Constructor string
}

type encodedLiquid struct {
	encodedBlock
	Still   bool
	Depth   int
	Falling bool
}

type liquidSpec struct {
	Name       string
	LiquidType string
	States     []encodedLiquid
}

type blockSpec struct {
	Types   []blockTypeSpec
	Liquids []liquidSpec
}

type encodedBiome struct {
	Name string
	ID   int
}

type particleType struct {
	Name   string
	Kind   uint32
	Fields []parameter
}

type instrumentSpec struct {
	Name string
	ID   uint32
}

type particleSpec struct {
	Types       []particleType
	Instruments []instrumentSpec
	RGBAFields  []parameter
}

type gameModeValue struct {
	Name         string
	PrivateType  string
	ID           int
	Capabilities []bool
}

type gameModeSpec struct {
	Methods []string
	Modes   []gameModeValue
}

type itemFieldKind uint8

const (
	itemFieldBool itemFieldKind = iota
	itemFieldToolTier
	itemFieldValue
)

type itemFieldSpec struct {
	Name      string
	Kind      itemFieldKind
	ValueType string
}

type itemStateSpec struct {
	Identifier string
	Metadata   int
	Bools      []bool
	ToolTier   int
	ArmourTier int
	Bucket     bucketContentKind
	Values     []int
	Capability itemCapabilitySpec
}

type itemCapabilitySpec struct {
	MaxCount         int
	MaxDurability    int
	Persistent       bool
	BrokenIdentifier string
	BrokenMetadata   int
	BrokenCount      int
	AttackDamage     float64
	AllowsAnvilCost  bool
	Fuel             bool
	FuelDuration     time.Duration
	FuelIdentifier   string
	FuelMetadata     int
	FuelCount        int
}

type itemTypeSpec struct {
	Name   string
	Fields []itemFieldSpec
	States []itemStateSpec
	NBT    bool
	Armour bool
	Bucket bool
}

type toolTierSpec struct {
	Variable string
	Value    dfitem.ToolTier
}

type itemSpec struct {
	Types          []itemTypeSpec
	ToolTiers      []toolTierSpec
	ToolTierFields []parameter
	ValueTypes     []itemValueTypeSpec
	Armour         armourSpec
	Crossbow       crossbowSpec
	Bucket         bucketSpec
	AirIdentifier  string
	Enchantments   []enchantmentSpec
}

type enchantmentSpec struct {
	Name                   string
	ID                     int
	DisplayName            string
	MaxLevel               int
	CompatibleEnchantments uint64
	CompatibleItems        []encodedItemKey
}

type encodedItemKey struct {
	Identifier string
	Metadata   int
}

type bucketContentKind uint8

const (
	bucketEmpty bucketContentKind = iota
	bucketWater
	bucketLava
	bucketMilk
)

type bucketSpec struct {
	Present               bool
	ConsumeDuration       time.Duration
	EmptyMaxCount         int
	FullMaxCount          int
	FuelDuration          time.Duration
	FuelResidueIdentifier string
	FuelResidueMetadata   int
	FuelResidueCount      int
}

type crossbowSpec struct {
	Present          bool
	MaxCount         int
	MaxDurability    int
	EnchantmentValue int
	FuelDuration     time.Duration
}

type armourSpec struct {
	Tiers         []armourTierSpec
	Pieces        []armourPieceSpec
	TrimMaterials []armourTrimMaterialSpec
}

type armourTierSpec struct {
	Name                string
	BaseDurability      float64
	Toughness           float64
	KnockBackResistance float64
	EnchantmentValue    int
	IdentifierName      string
	Colour              bool
}

type armourPieceSpec struct {
	Name              string
	SlotMethod        string
	DurabilityDivisor float64
	DefencePoints     []float64
	RepairItems       []string
	Smelts            []armourSmeltSpec
}

type armourSmeltSpec struct {
	Product    string
	Count      int
	Experience float64
	Food       bool
	Ores       bool
}

type armourTrimMaterialSpec struct {
	ItemName       string
	Material       string
	MaterialColour string
}

type itemValueTypeSpec struct {
	GoType     string
	CSharpType string
	Container  string
	Name       string
	Collection string
	From       bool
	Factories  []string
	Values     []any
	Methods    []itemValueMethodSpec
}

type itemValueMethodSpec struct {
	Name       string
	ReturnType string
	Results    []string
	Default    string
}

type formFieldSpec struct {
	Name string
	Type string
}

type formElementSpec struct {
	Name        string
	Fields      []formFieldSpec
	Element     bool
	MenuElement bool
	Constructor []parameter
	Tooltip     bool
	ValueType   string
}

type formSpec struct {
	Elements []formElementSpec
}

type goSignature struct {
	Parameters string
	Results    string
}

var gameModeMethodNames = []string{
	"AllowsEditing",
	"AllowsTakingDamage",
	"CreativeInventory",
	"HasCollision",
	"AllowsFlying",
	"AllowsInteraction",
	"Visible",
	"InstantPortalTravel",
}

var gameModeVariableNames = []string{
	"GameModeSurvival",
	"GameModeCreative",
	"GameModeAdventure",
	"GameModeSpectator",
}

var particleKindNames = []string{
	"Flame",
	"Dust",
	"BlockBreak",
	"PunchBlock",
	"BlockForceField",
	"BoneMeal",
	"Note",
	"DragonEggTeleport",
	"Evaporate",
	"WaterDrip",
	"LavaDrip",
	"Lava",
	"DustPlume",
	"HugeExplosion",
	"EndermanTeleport",
	"SnowballPoof",
	"EggSmash",
	"Splash",
	"Effect",
	"EntityFlame",
}

var instrumentNames = []string{
	"Piano",
	"BassDrum",
	"Snare",
	"ClicksAndSticks",
	"Bass",
	"Flute",
	"Bell",
	"Guitar",
	"Chimes",
	"Xylophone",
	"IronXylophone",
	"CowBell",
	"Didgeridoo",
	"Bit",
	"Banjo",
	"Pling",
}

var supportedPlayerHandlers = map[string]uint64{
	"HandleMove":             1,
	"HandleChat":             1 << 1,
	"HandleQuit":             1 << 3,
	"HandleHurt":             1 << 4,
	"HandleHeal":             1 << 5,
	"HandleBlockBreak":       1 << 6,
	"HandleBlockPlace":       1 << 7,
	"HandleFoodLoss":         1 << 8,
	"HandleDeath":            1 << 9,
	"HandleStartBreak":       1 << 10,
	"HandleFireExtinguish":   1 << 11,
	"HandleToggleSprint":     1 << 12,
	"HandleToggleSneak":      1 << 13,
	"HandleJump":             1 << 14,
	"HandleTeleport":         1 << 15,
	"HandleExperienceGain":   1 << 16,
	"HandlePunchAir":         1 << 17,
	"HandleHeldSlotChange":   1 << 18,
	"HandleSleep":            1 << 19,
	"HandleBlockPick":        1 << 20,
	"HandleLecternPageTurn":  1 << 21,
	"HandleSignEdit":         1 << 22,
	"HandleItemUse":          1 << 23,
	"HandleItemUseOnBlock":   1 << 24,
	"HandleItemConsume":      1 << 25,
	"HandleItemRelease":      1 << 26,
	"HandleItemDamage":       1 << 27,
	"HandleItemDrop":         1 << 28,
	"HandleAttackEntity":     1 << 29,
	"HandleItemUseOnEntity":  1 << 30,
	"HandleChangeWorld":      1 << 31,
	"HandleRespawn":          1 << 32,
	"HandleSkinChange":       1 << 33,
	"HandleItemPickup":       1 << 37,
	"HandleTransfer":         1 << 38,
	"HandleCommandExecution": 1 << 39,
	"HandleDiagnostics":      1 << 40,
}

var selectedCommandInterfaces = []string{
	"Runnable",
	"Allower",
	"Target",
	"NamedTarget",
	"Source",
	"Enum",
}

var selectedPlayerTextMethods = []string{
	"Message",
	"SendPopup",
	"SendTip",
	"SendJukeboxPopup",
	"SetNameTag",
	"Disconnect",
}

var selectedPlayerFormMethods = []string{
	"SendForm",
	"CloseForm",
}

var selectedFormElements = []string{
	"Divider",
	"Header",
	"Label",
	"Input",
	"Toggle",
	"Slider",
	"Dropdown",
	"StepSlider",
	"Button",
}

var selectedWorldTxMethods = []string{
	"World",
	"Range",
	"SetBlock",
	"Block",
	"BlockLoaded",
	"BlocksWithin",
	"Liquid",
	"SetLiquid",
	"ScheduleBlockUpdate",
	"HighestLightBlocker",
	"HighestBlock",
	"Light",
	"SkyLight",
	"SetBiome",
	"Biome",
	"Temperature",
	"RainingAt",
	"SnowingAt",
	"ThunderingAt",
	"Raining",
	"Thundering",
	"CurrentTick",
	"AddParticle",
	"PlaySound",
	"AddEntity",
	"AddEntityAt",
	"RemoveEntity",
	"Entities",
	"EntitiesWithin",
	"Players",
}

func main() {
	root := flag.String("root", ".", "repository root")
	dragonfly := flag.String("dragonfly", "", "Dragonfly module directory")
	gophertunnel := flag.String("gophertunnel", "", "gophertunnel module directory")
	intercept := flag.String("intercept", "", "bedrock-gophers/intercept module directory")
	unsafeModule := flag.String("unsafe", "", "bedrock-gophers/unsafe module directory")
	check := flag.Bool("check", false, "fail if generated output differs")
	flag.Parse()

	directory := *dragonfly
	if directory == "" {
		command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
		command.Dir = *root
		output, err := command.Output()
		if err != nil {
			fatal(err)
		}
		directory = string(bytes.TrimSpace(output))
	}
	gophertunnelDirectory := *gophertunnel
	if gophertunnelDirectory == "" {
		gophertunnelDirectory = moduleDirectory(*root, "github.com/sandertv/gophertunnel")
	}
	interceptDirectory := *intercept
	if interceptDirectory == "" {
		interceptDirectory = moduleDirectory(*root, "github.com/bedrock-gophers/intercept")
	}
	unsafeDirectory := *unsafeModule
	if unsafeDirectory == "" {
		unsafeDirectory = moduleDirectory(*root, "github.com/bedrock-gophers/unsafe")
	}
	packets, err := inspectPackets(filepath.Join(gophertunnelDirectory, "minecraft", "protocol", "packet"))
	if err != nil {
		fatal(err)
	}
	if err := inspectInterceptHandler(filepath.Join(interceptDirectory, "intercept", "handler.go")); err != nil {
		fatal(err)
	}
	if err := inspectUnsafeWritePacket(filepath.Join(unsafeDirectory, "unsafe.go")); err != nil {
		fatal(err)
	}
	methods, err := playerHandlerMethods(filepath.Join(directory, "server", "player", "handler.go"))
	if err != nil {
		fatal(err)
	}
	worldHandlerMethods, err := inspectWorldHandler(filepath.Join(directory, "server", "world", "handler.go"))
	if err != nil {
		fatal(err)
	}
	redstoneUpdate, err := inspectRedstoneUpdate(filepath.Join(directory, "server", "world", "redstone.go"))
	if err != nil {
		fatal(err)
	}
	entityMethods, err := inspectWorldEntity(filepath.Join(directory, "server", "world", "entity.go"))
	if err != nil {
		fatal(err)
	}
	entityConstruction, err := inspectWorldEntityConstruction(filepath.Join(directory, "server", "world", "entity.go"))
	if err != nil {
		fatal(err)
	}
	entityHandleMethods, err := inspectWorldEntityHandle(filepath.Join(directory, "server", "world", "entity.go"))
	if err != nil {
		fatal(err)
	}
	serverMethods, err := inspectServerMethods(filepath.Join(directory, "server", "server.go"))
	if err != nil {
		fatal(err)
	}
	serverAllower, err := inspectServerAllower(filepath.Join(directory, "server", "allower.go"))
	if err != nil {
		fatal(err)
	}
	loginData, err := inspectLoginData(filepath.Join(gophertunnelDirectory, "minecraft", "protocol", "login", "data.go"))
	if err != nil {
		fatal(err)
	}
	deviceOS, err := inspectDeviceOS(filepath.Join(gophertunnelDirectory, "minecraft", "protocol", "os.go"))
	if err != nil {
		fatal(err)
	}
	playerIdentityMethods, err := inspectPlayerIdentityMethods(filepath.Join(directory, "server", "player", "player.go"))
	if err != nil {
		fatal(err)
	}
	playerMethods, err := playerTextMethods(filepath.Join(directory, "server", "player", "player.go"))
	if err != nil {
		fatal(err)
	}
	playerFormMethods, err := inspectPlayerFormMethods(filepath.Join(directory, "server", "player", "player.go"))
	if err != nil {
		fatal(err)
	}
	playerItemMethods, err := inspectPlayerItemMethods(filepath.Join(directory, "server", "player", "player.go"))
	if err != nil {
		fatal(err)
	}
	playerStateMethods, err := inspectPlayerStateMethods(filepath.Join(directory, "server", "player", "player.go"))
	if err != nil {
		fatal(err)
	}
	playerActionMethods, err := inspectPlayerActionMethods(filepath.Join(directory, "server", "player", "player.go"))
	if err != nil {
		fatal(err)
	}
	playerControls, err := inspectPlayerControls(
		filepath.Join(directory, "server", "player", "player.go"),
		filepath.Join(directory, "server", "player", "hud", "element.go"),
		filepath.Join(directory, "server", "player", "input", "lock.go"),
		filepath.Join(gophertunnelDirectory, "minecraft", "protocol", "packet", "update_client_input_locks.go"),
	)
	if err != nil {
		fatal(err)
	}
	playerPresentationMethods, err := inspectPlayerPresentationMethods(filepath.Join(directory, "server", "player", "player.go"))
	if err != nil {
		fatal(err)
	}
	playerCooldown, err := inspectPlayerCooldown(filepath.Join(directory, "server", "player", "player.go"))
	if err != nil {
		fatal(err)
	}
	playerFinalDamage, err := inspectPlayerFinalDamage(filepath.Join(directory, "server", "player", "player.go"))
	if err != nil {
		fatal(err)
	}
	playerSkin, err := inspectPlayerSkin(filepath.Join(directory, "server", "player", "player.go"))
	if err != nil {
		fatal(err)
	}
	playerVisibility, err := inspectPlayerVisibility(filepath.Join(directory, "server", "player", "player.go"))
	if err != nil {
		fatal(err)
	}
	if err := inspectTitle(
		filepath.Join(directory, "server", "player", "title", "title.go"),
		filepath.Join(directory, "server", "player", "player.go"),
	); err != nil {
		fatal(err)
	}
	if err := inspectScoreboard(
		filepath.Join(directory, "server", "player", "scoreboard", "scoreboard.go"),
		filepath.Join(directory, "server", "player", "player.go"),
	); err != nil {
		fatal(err)
	}
	playerKinematicsMethods, err := inspectPlayerKinematicsMethods(filepath.Join(directory, "server", "player", "player.go"))
	if err != nil {
		fatal(err)
	}
	worldLifecycleMethods, err := inspectWorldLifecycleMethods(filepath.Join(directory, "server", "world", "world.go"))
	if err != nil {
		fatal(err)
	}
	if err := inspectWorldSchedule(filepath.Join(directory, "server", "world", "task.go")); err != nil {
		fatal(err)
	}
	if err := inspectWorldConfig(
		filepath.Join(directory, "server", "world", "conf.go"),
		filepath.Join(directory, "server", "world", "world.go"),
		filepath.Join(directory, "server", "world", "dimension.go"),
		filepath.Join(directory, "server", "world", "provider.go"),
		filepath.Join(directory, "server", "world", "mcdb", "conf.go"),
	); err != nil {
		fatal(err)
	}
	effects, err := inspectEffects(filepath.Join(directory, "server", "entity", "effect"))
	if err != nil {
		fatal(err)
	}
	sourceTypes, err := inspectSourceTypes(
		filepath.Join(directory, "server", "world"),
		filepath.Join(directory, "server", "entity"),
		filepath.Join(directory, "server", "entity", "effect"),
		filepath.Join(directory, "server", "player"),
		filepath.Join(directory, "server", "block"),
		filepath.Join(directory, "server", "item", "enchantment"),
	)
	if err != nil {
		fatal(err)
	}
	playerEffectMethods, err := inspectPlayerEffectMethods(filepath.Join(directory, "server", "player", "player.go"))
	if err != nil {
		fatal(err)
	}
	forms, err := inspectForms(filepath.Join(directory, "server", "player", "form"))
	if err != nil {
		fatal(err)
	}
	interfaces, err := commandInterfaces(filepath.Join(directory, "server", "cmd"))
	if err != nil {
		fatal(err)
	}
	cube, err := inspectCube(filepath.Join(directory, "server", "block", "cube"))
	if err != nil {
		fatal(err)
	}
	bbox, err := inspectBBox(
		filepath.Join(directory, "server", "block", "cube", "bbox.go"),
		filepath.Join(directory, "server", "block", "cube", "axis.go"),
	)
	if err != nil {
		fatal(err)
	}
	if err := inspectWorldBlockByName(filepath.Join(directory, "server", "world", "block.go")); err != nil {
		fatal(err)
	}
	setOpts, err := inspectSetOpts(filepath.Join(directory, "server", "world", "world.go"))
	if err != nil {
		fatal(err)
	}
	worldTx, err := inspectWorldTx(filepath.Join(directory, "server", "world", "tx.go"))
	if err != nil {
		fatal(err)
	}
	blocks, err := inspectBlocks(filepath.Join(directory, "server", "block"))
	if err != nil {
		fatal(err)
	}
	biomes, err := inspectBiomes(filepath.Join(directory, "server", "world", "biome"))
	if err != nil {
		fatal(err)
	}
	items, err := inspectItems(filepath.Join(directory, "server", "item"), effects)
	if err != nil {
		fatal(err)
	}
	itemStack, err := inspectItemStack(filepath.Join(directory, "server", "item", "stack.go"))
	if err != nil {
		fatal(err)
	}
	inventories, err := inspectInventories(filepath.Join(directory, "server", "item", "inventory"))
	if err != nil {
		fatal(err)
	}
	particles, err := inspectParticles(
		filepath.Join(directory, "server", "world", "particle"),
		filepath.Join(directory, "server", "world", "sound", "instrument.go"),
		filepath.Join(runtime.GOROOT(), "src", "image", "color", "color.go"),
	)
	if err != nil {
		fatal(err)
	}
	sounds, err := inspectSounds(filepath.Join(directory, "server", "world", "sound"))
	if err != nil {
		fatal(err)
	}
	playerPlaySound, err := inspectPlayerPlaySound(filepath.Join(directory, "server", "player", "player.go"))
	if err != nil {
		fatal(err)
	}
	gameModes, err := inspectGameModes(filepath.Join(directory, "server", "world", "game_mode.go"))
	if err != nil {
		fatal(err)
	}
	difficulties, err := inspectDifficulties(filepath.Join(directory, "server", "world", "difficulty.go"))
	if err != nil {
		fatal(err)
	}
	playerGameModes, err := inspectPlayerGameModeMethods(filepath.Join(directory, "server", "player", "player.go"))
	if err != nil {
		fatal(err)
	}
	playerTransport := playerTransportSpec{
		StateMethods: playerStateMethods, ActionMethods: playerActionMethods, Controls: playerControls, IdentityMethods: playerIdentityMethods, PresentationMethods: playerPresentationMethods,
		TextMethods: playerMethods, Effects: effects, Sounds: sounds, GameModeMethods: playerGameModes,
	}
	nativePlayerTransport, err := generateNativePlayerTransport(playerTransport)
	if err != nil {
		fatal(err)
	}
	csharpPlayerStateTransport, err := generateCSharpPlayerStateTransport(playerTransport)
	if err != nil {
		fatal(err)
	}
	hostPlayerTransport, err := generateHostPlayerTransport(playerTransport)
	if err != nil {
		fatal(err)
	}
	files := []generatedFile{
		{
			Path:    filepath.Join(*root, "internal", "native", "player_state_generated.go"),
			Content: nativePlayerTransport,
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly.Native", "Generated", "Player.State.g.cs"),
			Content: csharpPlayerStateTransport,
		},
		{
			Path:    filepath.Join(*root, "internal", "host", "player_state_generated.go"),
			Content: hostPlayerTransport,
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Packet.Types.g.cs"),
			Content: generatePacketTypes(packets),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Packet.Handler.g.cs"),
			Content: generatePacketHandler(),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Packet.g.cs"),
			Content: generatePlayerWritePacket(),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Server.g.cs"),
			Content: generateServer(serverMethods),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Server.Allower.g.cs"),
			Content: generateServerAllower(serverAllower, loginData, deviceOS),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Identity.g.cs"),
			Content: generatePlayerIdentityMethods(playerIdentityMethods),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "World.Entity.g.cs"),
			Content: generateWorldEntity(entityMethods, entityHandleMethods, entityConstruction),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Handler.g.cs"),
			Content: generatePlayerHandler(methods),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "World.Handler.g.cs"),
			Content: generateWorldHandler(worldHandlerMethods, redstoneUpdate),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Text.g.cs"),
			Content: generatePlayerTextMethods(playerMethods),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Form.g.cs"),
			Content: generatePlayerFormMethods(playerFormMethods),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Item.g.cs"),
			Content: generatePlayerItemMethods(playerItemMethods),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.State.g.cs"),
			Content: generatePlayerStateMethods(playerStateMethods),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Action.g.cs"),
			Content: generatePlayerActionMethods(playerActionMethods),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Control.g.cs"),
			Content: generatePlayerControls(playerControls),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Presentation.g.cs"),
			Content: generatePlayerPresentationMethods(playerPresentationMethods),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Title.g.cs"),
			Content: generateTitle(),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Scoreboard.g.cs"),
			Content: generateScoreboard(),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Cooldown.g.cs"),
			Content: generatePlayerCooldown(playerCooldown),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.FinalDamage.g.cs"),
			Content: generatePlayerFinalDamage(playerFinalDamage),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Skin.g.cs"),
			Content: generatePlayerSkin(playerSkin),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Visibility.g.cs"),
			Content: generatePlayerVisibility(playerVisibility),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Kinematics.g.cs"),
			Content: generatePlayerKinematicsMethods(playerKinematicsMethods),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "World.Lifecycle.g.cs"),
			Content: generateWorldLifecycleMethods(worldLifecycleMethods),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "World.Schedule.g.cs"),
			Content: generateWorldSchedule(),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "World.Config.g.cs"),
			Content: generateWorldConfig(),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Effect.Types.g.cs"),
			Content: generateEffects(effects),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "World.Source.g.cs"),
			Content: generateSourceTypes(sourceTypes),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Effect.g.cs"),
			Content: generatePlayerEffects(playerEffectMethods),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Form.Types.g.cs"),
			Content: generateForms(forms),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Cmd.Interfaces.g.cs"),
			Content: generateCommandInterfaces(interfaces),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Cube.g.cs"),
			Content: generateCube(cube),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Cube.BBox.g.cs"),
			Content: generateBBox(bbox),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "World.Block.g.cs"),
			Content: generateWorldBlock(setOpts, worldTx),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Block.Types.g.cs"),
			Content: generateBlocks(blocks),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Biome.Types.g.cs"),
			Content: generateBiomes(biomes),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Item.Types.g.cs"),
			Content: generateItems(items),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Item.Stack.g.cs"),
			Content: generateItemStack(itemStack),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Inventory.g.cs"),
			Content: generateInventories(inventories),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Particle.Types.g.cs"),
			Content: generateParticles(particles),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Sound.Types.g.cs"),
			Content: generateSounds(sounds),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Sound.g.cs"),
			Content: generatePlayerPlaySound(playerPlaySound),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "GameMode.Types.g.cs"),
			Content: generateGameModes(gameModes),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Difficulty.Types.g.cs"),
			Content: generateDifficulties(difficulties),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.GameMode.g.cs"),
			Content: generatePlayerGameModes(playerGameModes),
		},
	}
	if err := syncGeneratedFiles(files, *check); err != nil {
		fatal(err)
	}
}

func moduleDirectory(root, module string) string {
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", module)
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		fatal(fmt.Errorf("locate %s: %w", module, err))
	}
	return string(bytes.TrimSpace(output))
}

func playerTextMethods(path string) ([]method, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	found := map[string]method{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !selectedPlayerTextMethod(function.Name.Name) || !playerMethod(function) {
			continue
		}
		if function.Type.Results != nil && len(function.Type.Results.List) != 0 {
			return nil, fmt.Errorf("player.Player.%s returns values", function.Name.Name)
		}
		parameters, err := translateParameters(function.Type.Params)
		if err != nil {
			return nil, fmt.Errorf("player.Player.%s: %w", function.Name.Name, err)
		}
		found[function.Name.Name] = method{Name: function.Name.Name, Parameters: parameters}
	}
	methods := make([]method, 0, len(selectedPlayerTextMethods))
	for _, name := range selectedPlayerTextMethods {
		definition, ok := found[name]
		if !ok {
			return nil, fmt.Errorf("Dragonfly player.Player has no supported %s method", name)
		}
		methods = append(methods, definition)
	}
	return methods, nil
}

func selectedPlayerTextMethod(name string) bool {
	for _, selected := range selectedPlayerTextMethods {
		if name == selected {
			return true
		}
	}
	return false
}

func inspectPlayerFormMethods(path string) ([]method, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	found := map[string]method{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !playerMethod(function) {
			continue
		}
		selected := false
		for _, name := range selectedPlayerFormMethods {
			selected = selected || function.Name.Name == name
		}
		if !selected {
			continue
		}
		if function.Type.Results != nil && len(function.Type.Results.List) != 0 {
			return nil, fmt.Errorf("player.Player.%s returns values", function.Name.Name)
		}
		parameters, err := translateFormParameters(function.Type.Params)
		if err != nil {
			return nil, fmt.Errorf("player.Player.%s: %w", function.Name.Name, err)
		}
		found[function.Name.Name] = method{Name: function.Name.Name, Parameters: parameters}
	}
	methods := make([]method, 0, len(selectedPlayerFormMethods))
	for _, name := range selectedPlayerFormMethods {
		definition, ok := found[name]
		if !ok {
			return nil, fmt.Errorf("Dragonfly player.Player has no supported %s method", name)
		}
		methods = append(methods, definition)
	}
	return methods, nil
}

func inspectForms(directory string) (formSpec, error) {
	packages, err := parser.ParseDir(token.NewFileSet(), directory, nil, 0)
	if err != nil {
		return formSpec{}, err
	}
	pkg, ok := packages["form"]
	if !ok {
		return formSpec{}, fmt.Errorf("Dragonfly form package not found")
	}
	types := map[string]*ast.TypeSpec{}
	functions := map[string]*ast.FuncDecl{}
	methods := map[string]map[string]*ast.FuncDecl{}
	markers := map[string]map[string]bool{}
	for _, file := range pkg.Files {
		for _, declaration := range file.Decls {
			switch value := declaration.(type) {
			case *ast.GenDecl:
				for _, raw := range value.Specs {
					if spec, ok := raw.(*ast.TypeSpec); ok {
						types[spec.Name.Name] = spec
					}
				}
			case *ast.FuncDecl:
				if value.Recv == nil {
					functions[value.Name.Name] = value
					continue
				}
				receiver, ok := receiverName(value)
				if !ok {
					continue
				}
				if methods[receiver] == nil {
					methods[receiver] = map[string]*ast.FuncDecl{}
				}
				methods[receiver][value.Name.Name] = value
				if value.Name.Name == "elem" || value.Name.Name == "menuElem" {
					if markers[receiver] == nil {
						markers[receiver] = map[string]bool{}
					}
					markers[receiver][value.Name.Name] = true
				}
			}
		}
	}

	if err := validateFormInterfaces(types); err != nil {
		return formSpec{}, err
	}
	if err := validateFormContainers(types, functions, methods); err != nil {
		return formSpec{}, err
	}

	definitions := make([]formElementSpec, 0, len(selectedFormElements))
	for _, name := range selectedFormElements {
		typeSpec, ok := types[name]
		if !ok {
			return formSpec{}, fmt.Errorf("form.%s not found", name)
		}
		fields, err := formElementFields(name, typeSpec, types)
		if err != nil {
			return formSpec{}, err
		}
		definition := formElementSpec{
			Name: name, Fields: fields,
			Element: markers[name]["elem"], MenuElement: markers[name]["menuElem"],
		}
		marshal, ok := methods[name]["MarshalJSON"]
		if !ok || !valueReceiver(marshal, name) || rawParameterTypes(marshal.Type.Params) != "" || rawResultTypes(marshal.Type.Results) != "[]byte,error" {
			return formSpec{}, fmt.Errorf("form.%s.MarshalJSON signature changed", name)
		}
		constructorName := "New" + name
		if name != "Divider" {
			constructor, ok := functions[constructorName]
			if !ok {
				return formSpec{}, fmt.Errorf("form.%s not found", constructorName)
			}
			if err := requireSingleResult(constructor, name); err != nil {
				return formSpec{}, fmt.Errorf("form.%s: %w", constructorName, err)
			}
			definition.Constructor, err = translateFormParameters(constructor.Type.Params)
			if err != nil {
				return formSpec{}, fmt.Errorf("form.%s: %w", constructorName, err)
			}
		}
		if tooltip := methods[name]["WithTooltip"]; tooltip != nil {
			if !valueReceiver(tooltip, name) || requireMethodSignature(tooltip, []string{"string"}, name) != nil {
				return formSpec{}, fmt.Errorf("form.%s.WithTooltip signature changed", name)
			}
			definition.Tooltip = true
		}
		if value := methods[name]["Value"]; value != nil {
			if !valueReceiver(value, name) || value.Type.Params.NumFields() != 0 || value.Type.Results == nil || len(value.Type.Results.List) != 1 {
				return formSpec{}, fmt.Errorf("form.%s.Value signature changed", name)
			}
			translated, ok := formCSharpType(value.Type.Results.List[0].Type)
			if !ok {
				return formSpec{}, fmt.Errorf("form.%s.Value has unsupported result %s", name, formatGoExpression(value.Type.Results.List[0].Type))
			}
			definition.ValueType = translated
		}
		definitions = append(definitions, definition)
	}
	return formSpec{Elements: definitions}, nil
}

func receiverName(function *ast.FuncDecl) (string, bool) {
	if function.Recv == nil || len(function.Recv.List) != 1 {
		return "", false
	}
	expression := function.Recv.List[0].Type
	if pointer, ok := expression.(*ast.StarExpr); ok {
		expression = pointer.X
	}
	identifier, ok := expression.(*ast.Ident)
	return identifier.Name, ok
}

func valueReceiver(function *ast.FuncDecl, name string) bool {
	if function.Recv == nil || len(function.Recv.List) != 1 {
		return false
	}
	identifier, ok := function.Recv.List[0].Type.(*ast.Ident)
	return ok && identifier.Name == name
}

func formElementFields(name string, spec *ast.TypeSpec, types map[string]*ast.TypeSpec) ([]formFieldSpec, error) {
	if identifier, ok := spec.Type.(*ast.Ident); ok {
		underlying, found := types[identifier.Name]
		if !found {
			return nil, fmt.Errorf("form.%s has unknown underlying type %s", name, identifier.Name)
		}
		return formElementFields(name, underlying, types)
	}
	structure, ok := spec.Type.(*ast.StructType)
	if !ok {
		return nil, fmt.Errorf("form.%s is not a struct", name)
	}
	var fields []formFieldSpec
	for _, field := range structure.Fields.List {
		typeName, ok := formCSharpType(field.Type)
		if !ok {
			return nil, fmt.Errorf("form.%s has unsupported field type %s", name, formatGoExpression(field.Type))
		}
		for _, fieldName := range field.Names {
			if fieldName.IsExported() {
				fields = append(fields, formFieldSpec{Name: fieldName.Name, Type: typeName})
			}
		}
	}
	return fields, nil
}

func validateFormInterfaces(types map[string]*ast.TypeSpec) error {
	want := map[string]map[string]goSignature{
		"Submittable":     {"Submit": {Parameters: "Submitter,*world.Tx"}},
		"MenuSubmittable": {"Submit": {Parameters: "Submitter,Button,*world.Tx"}},
		"Closer":          {"Close": {Parameters: "Submitter,*world.Tx"}},
		"Submitter":       {"SendForm": {Parameters: "Form"}, "CloseForm": {}},
		"Form":            {"SubmitJSON": {Parameters: "[]byte,Submitter,*world.Tx", Results: "error"}},
	}
	for name, methods := range want {
		typeSpec, ok := types[name]
		if !ok {
			return fmt.Errorf("form.%s interface not found", name)
		}
		interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
		if !ok {
			return fmt.Errorf("form.%s is not an interface", name)
		}
		found := map[string]goSignature{}
		for _, field := range interfaceType.Methods.List {
			if len(field.Names) != 1 {
				continue
			}
			function, ok := field.Type.(*ast.FuncType)
			if !ok {
				return fmt.Errorf("form.%s.%s is not a method", name, field.Names[0].Name)
			}
			found[field.Names[0].Name] = goSignature{
				Parameters: rawParameterTypes(function.Params),
				Results:    rawResultTypes(function.Results),
			}
		}
		if !reflect.DeepEqual(found, methods) {
			return fmt.Errorf("form.%s methods changed: %v", name, found)
		}
	}
	for name, marker := range map[string]string{"Element": "elem", "MenuElement": "menuElem"} {
		typeSpec, ok := types[name]
		if !ok {
			return fmt.Errorf("form.%s interface not found", name)
		}
		interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
		if !ok || len(interfaceType.Methods.List) != 2 {
			return fmt.Errorf("form.%s interface changed", name)
		}
		var marshaler, markerMethod bool
		for _, field := range interfaceType.Methods.List {
			if len(field.Names) == 0 {
				marshaler = formatGoExpression(field.Type) == "json.Marshaler"
				continue
			}
			function, ok := field.Type.(*ast.FuncType)
			markerMethod = ok && field.Names[0].Name == marker && rawParameterTypes(function.Params) == "" && rawResultTypes(function.Results) == ""
		}
		if !marshaler || !markerMethod {
			return fmt.Errorf("form.%s interface changed", name)
		}
	}
	modal, ok := types["ModalSubmittable"]
	if !ok {
		return fmt.Errorf("form.ModalSubmittable not found")
	}
	identifier, ok := modal.Type.(*ast.Ident)
	if !ok || identifier.Name != "MenuSubmittable" {
		return fmt.Errorf("form.ModalSubmittable underlying type changed")
	}
	return nil
}

func validateFormContainers(types map[string]*ast.TypeSpec, functions map[string]*ast.FuncDecl, methods map[string]map[string]*ast.FuncDecl) error {
	constructors := map[string]struct {
		Result string
		Params string
	}{
		"New":      {"Custom", "Submittable,...any"},
		"NewMenu":  {"Menu", "MenuSubmittable,...any"},
		"NewModal": {"Modal", "ModalSubmittable,...any"},
	}
	for name, expected := range constructors {
		function, ok := functions[name]
		if !ok || rawParameterTypes(function.Type.Params) != expected.Params || requireSingleResult(function, expected.Result) != nil {
			return fmt.Errorf("form.%s signature changed", name)
		}
	}
	for _, name := range []string{"Custom", "Menu", "Modal"} {
		if _, ok := types[name]; !ok {
			return fmt.Errorf("form.%s not found", name)
		}
	}
	wantMethods := map[string]map[string]goSignature{
		"Custom": {
			"Title": {Results: "string"}, "Elements": {Results: "[]Element"},
			"SubmitJSON": {Parameters: "[]byte,Submitter,*world.Tx", Results: "error"}, "MarshalJSON": {Results: "[]byte,error"},
		},
		"Menu": {
			"WithBody": {Parameters: "...any", Results: "Menu"}, "AddButton": {Parameters: "Button", Results: "Menu"},
			"AddDivider": {Parameters: "Divider", Results: "Menu"}, "AddHeader": {Parameters: "Header", Results: "Menu"},
			"AddLabel": {Parameters: "Label", Results: "Menu"}, "WithButtons": {Parameters: "...Button", Results: "Menu"},
			"WithElements": {Parameters: "...MenuElement", Results: "Menu"}, "Title": {Results: "string"}, "Body": {Results: "string"},
			"Buttons": {Results: "[]Button"}, "Elements": {Results: "[]MenuElement"},
			"SubmitJSON": {Parameters: "[]byte,Submitter,*world.Tx", Results: "error"}, "MarshalJSON": {Results: "[]byte,error"},
		},
		"Modal": {
			"WithBody": {Parameters: "...any", Results: "Modal"}, "Title": {Results: "string"}, "Body": {Results: "string"},
			"Buttons": {Results: "[]Button"}, "SubmitJSON": {Parameters: "[]byte,Submitter,*world.Tx", Results: "error"},
			"MarshalJSON": {Results: "[]byte,error"},
		},
	}
	for receiver, expected := range wantMethods {
		for name, signature := range expected {
			function, ok := methods[receiver][name]
			if !ok || !valueReceiver(function, receiver) || rawParameterTypes(function.Type.Params) != signature.Parameters || rawResultTypes(function.Type.Results) != signature.Results {
				return fmt.Errorf("form.%s.%s signature changed", receiver, name)
			}
		}
	}
	for _, name := range []string{"YesButton", "NoButton"} {
		function, ok := functions[name]
		if !ok || rawParameterTypes(function.Type.Params) != "" || requireSingleResult(function, "Button") != nil {
			return fmt.Errorf("form.%s signature changed", name)
		}
	}
	return nil
}

func rawParameterTypes(fields *ast.FieldList) string {
	if fields == nil {
		return ""
	}
	var values []string
	for _, field := range fields.List {
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for range count {
			values = append(values, formatGoExpression(field.Type))
		}
	}
	return strings.Join(values, ",")
}

func rawResultTypes(fields *ast.FieldList) string {
	return rawParameterTypes(fields)
}

func requireSingleResult(function *ast.FuncDecl, result string) error {
	if function.Type.Results == nil || len(function.Type.Results.List) != 1 || len(function.Type.Results.List[0].Names) > 1 || formatGoExpression(function.Type.Results.List[0].Type) != result {
		return fmt.Errorf("must return %s", result)
	}
	return nil
}

func requireMethodSignature(function *ast.FuncDecl, parameters []string, result string) error {
	if rawParameterTypes(function.Type.Params) != strings.Join(parameters, ",") {
		return fmt.Errorf("parameters changed")
	}
	return requireSingleResult(function, result)
}

func translateFormParameters(fields *ast.FieldList) ([]parameter, error) {
	var parameters []parameter
	if fields == nil {
		return parameters, nil
	}
	for _, field := range fields.List {
		typeName, ok := formCSharpType(field.Type)
		if !ok {
			return nil, fmt.Errorf("unsupported parameter type %s", formatGoExpression(field.Type))
		}
		for _, name := range field.Names {
			parameters = append(parameters, parameter{Name: name.Name, Type: typeName})
		}
	}
	return parameters, nil
}

func formCSharpType(expression ast.Expr) (string, bool) {
	switch value := expression.(type) {
	case *ast.Ellipsis:
		typeName, ok := formCSharpType(value.Elt)
		return "params " + typeName + "[]", ok
	case *ast.ArrayType:
		if value.Len != nil {
			return "", false
		}
		element, ok := formCSharpType(value.Elt)
		return element + "[]", ok
	case *ast.Ident:
		translated, ok := map[string]string{
			"any": "object?", "bool": "bool", "float64": "double", "int": "int", "string": "string", "Form": "Form.Value",
		}[value.Name]
		return translated, ok
	case *ast.SelectorExpr:
		packageName, ok := value.X.(*ast.Ident)
		if !ok {
			return "", false
		}
		translated, ok := map[string]string{"form.Form": "Form.Value"}[packageName.Name+"."+value.Sel.Name]
		return translated, ok
	default:
		return "", false
	}
}

func playerMethod(function *ast.FuncDecl) bool {
	if function.Recv == nil || len(function.Recv.List) != 1 {
		return false
	}
	pointer, ok := function.Recv.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	receiver, ok := pointer.X.(*ast.Ident)
	return ok && receiver.Name == "Player"
}

func inspectCube(directory string) (cubeSpec, error) {
	if err := requireArrayType(filepath.Join(directory, "pos.go"), "Pos", 3, "int"); err != nil {
		return cubeSpec{}, err
	}
	if err := requireArrayType(filepath.Join(directory, "range.go"), "Range", 2, "int"); err != nil {
		return cubeSpec{}, err
	}
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(directory, "face.go"), nil, 0)
	if err != nil {
		return cubeSpec{}, err
	}
	if !hasNamedType(file, "Face", "int") {
		return cubeSpec{}, fmt.Errorf("cube.Face is not backed by int")
	}
	var faces []string
	for _, declaration := range file.Decls {
		gen, ok := declaration.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, raw := range gen.Specs {
			spec, ok := raw.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range spec.Names {
				if strings.HasPrefix(name.Name, "Face") {
					faces = append(faces, strings.TrimPrefix(name.Name, "Face"))
				}
			}
		}
	}
	want := []string{"Down", "Up", "North", "South", "West", "East"}
	if !reflect.DeepEqual(faces, want) {
		return cubeSpec{}, fmt.Errorf("cube.Face values changed: %v", faces)
	}
	return cubeSpec{Faces: faces}, nil
}

func requireArrayType(path, name string, length int64, element string) error {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return err
	}
	for _, declaration := range file.Decls {
		gen, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, raw := range gen.Specs {
			spec, ok := raw.(*ast.TypeSpec)
			if !ok || spec.Name.Name != name {
				continue
			}
			array, ok := spec.Type.(*ast.ArrayType)
			if !ok || array.Len == nil {
				return fmt.Errorf("cube.%s is not a fixed array", name)
			}
			literal, ok := array.Len.(*ast.BasicLit)
			if !ok || literal.Kind != token.INT || literal.Value != strconv.FormatInt(length, 10) {
				return fmt.Errorf("cube.%s length changed", name)
			}
			identifier, ok := array.Elt.(*ast.Ident)
			if !ok || identifier.Name != element {
				return fmt.Errorf("cube.%s element changed", name)
			}
			return nil
		}
	}
	return fmt.Errorf("cube.%s not found", name)
}

func hasNamedType(file *ast.File, name, underlying string) bool {
	for _, declaration := range file.Decls {
		gen, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, raw := range gen.Specs {
			spec, ok := raw.(*ast.TypeSpec)
			if !ok || spec.Name.Name != name {
				continue
			}
			identifier, ok := spec.Type.(*ast.Ident)
			return ok && identifier.Name == underlying
		}
	}
	return false
}

func inspectSetOpts(path string) ([]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	for _, declaration := range file.Decls {
		gen, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, raw := range gen.Specs {
			spec, ok := raw.(*ast.TypeSpec)
			if !ok || spec.Name.Name != "SetOpts" {
				continue
			}
			structure, ok := spec.Type.(*ast.StructType)
			if !ok {
				return nil, fmt.Errorf("world.SetOpts is not a struct")
			}
			var fields []string
			for _, field := range structure.Fields.List {
				identifier, ok := field.Type.(*ast.Ident)
				if !ok || identifier.Name != "bool" {
					return nil, fmt.Errorf("world.SetOpts contains a non-bool field")
				}
				for _, name := range field.Names {
					if name.IsExported() {
						fields = append(fields, name.Name)
					}
				}
			}
			if len(fields) == 0 {
				return nil, fmt.Errorf("world.SetOpts has no exported fields")
			}
			return fields, nil
		}
	}
	return nil, fmt.Errorf("world.SetOpts not found")
}

func inspectWorldTx(path string) ([]commandMethod, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	found := map[string]commandMethod{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !selectedWorldTxMethod(function.Name.Name) || !pointerReceiver(function, "Tx") {
			continue
		}
		parameters, err := translateWorldTxParameters(function.Name.Name, function.Type.Params)
		if err != nil {
			return nil, fmt.Errorf("world.Tx.%s: %w", function.Name.Name, err)
		}
		result, err := translateWorldTxResult(function.Name.Name, function.Type.Results)
		if err != nil {
			return nil, fmt.Errorf("world.Tx.%s: %w", function.Name.Name, err)
		}
		found[function.Name.Name] = commandMethod{
			Name: function.Name.Name, Parameters: parameters, ReturnType: result,
		}
		if err := validateWorldTxMethod(found[function.Name.Name]); err != nil {
			return nil, fmt.Errorf("world.Tx.%s: %w", function.Name.Name, err)
		}
	}
	methods := make([]commandMethod, 0, len(selectedWorldTxMethods))
	for _, name := range selectedWorldTxMethods {
		definition, ok := found[name]
		if !ok {
			return nil, fmt.Errorf("Dragonfly world.Tx has no supported %s method", name)
		}
		methods = append(methods, definition)
	}
	return methods, nil
}

func validateWorldTxMethod(method commandMethod) error {
	expected := map[string]commandMethod{
		"World": {Name: "World", ReturnType: "World"},
		"Range": {Name: "Range", ReturnType: "Cube.Range"},
		"SetBlock": {Name: "SetBlock", ReturnType: "void", Parameters: []parameter{
			{Name: "pos", Type: "Cube.Pos"}, {Name: "b", Type: "Block?"}, {Name: "opts", Type: "SetOpts?"},
		}},
		"Block": {Name: "Block", ReturnType: "Block", Parameters: []parameter{
			{Name: "pos", Type: "Cube.Pos"},
		}},
		"BlockLoaded": {Name: "BlockLoaded", ReturnType: "(Block? Block, bool Ok)", Parameters: []parameter{
			{Name: "pos", Type: "Cube.Pos"},
		}},
		"BlocksWithin": {Name: "BlocksWithin", ReturnType: "IEnumerable<Cube.Pos>", Parameters: []parameter{
			{Name: "pos", Type: "Cube.Pos"}, {Name: "radius", Type: "int"}, {Name: "blocks", Type: "params Block[]"},
		}},
		"Liquid": {Name: "Liquid", ReturnType: "(Liquid? Liquid, bool Ok)", Parameters: []parameter{
			{Name: "pos", Type: "Cube.Pos"},
		}},
		"SetLiquid": {Name: "SetLiquid", ReturnType: "void", Parameters: []parameter{
			{Name: "pos", Type: "Cube.Pos"}, {Name: "b", Type: "Liquid?"},
		}},
		"ScheduleBlockUpdate": {Name: "ScheduleBlockUpdate", ReturnType: "void", Parameters: []parameter{
			{Name: "pos", Type: "Cube.Pos"}, {Name: "b", Type: "Block"}, {Name: "delay", Type: "TimeSpan"},
		}},
		"HighestLightBlocker": {Name: "HighestLightBlocker", ReturnType: "int", Parameters: []parameter{
			{Name: "x", Type: "int"}, {Name: "z", Type: "int"},
		}},
		"HighestBlock": {Name: "HighestBlock", ReturnType: "int", Parameters: []parameter{
			{Name: "x", Type: "int"}, {Name: "z", Type: "int"},
		}},
		"Light": {Name: "Light", ReturnType: "byte", Parameters: []parameter{
			{Name: "pos", Type: "Cube.Pos"},
		}},
		"SkyLight": {Name: "SkyLight", ReturnType: "byte", Parameters: []parameter{
			{Name: "pos", Type: "Cube.Pos"},
		}},
		"SetBiome": {Name: "SetBiome", ReturnType: "void", Parameters: []parameter{
			{Name: "pos", Type: "Cube.Pos"}, {Name: "b", Type: "Biome"},
		}},
		"Biome": {Name: "Biome", ReturnType: "Biome", Parameters: []parameter{
			{Name: "pos", Type: "Cube.Pos"},
		}},
		"Temperature": {Name: "Temperature", ReturnType: "double", Parameters: []parameter{
			{Name: "pos", Type: "Cube.Pos"},
		}},
		"RainingAt": {Name: "RainingAt", ReturnType: "bool", Parameters: []parameter{
			{Name: "pos", Type: "Cube.Pos"},
		}},
		"SnowingAt": {Name: "SnowingAt", ReturnType: "bool", Parameters: []parameter{
			{Name: "pos", Type: "Cube.Pos"},
		}},
		"ThunderingAt": {Name: "ThunderingAt", ReturnType: "bool", Parameters: []parameter{
			{Name: "pos", Type: "Cube.Pos"},
		}},
		"Raining":     {Name: "Raining", ReturnType: "bool"},
		"Thundering":  {Name: "Thundering", ReturnType: "bool"},
		"CurrentTick": {Name: "CurrentTick", ReturnType: "long"},
		"AddParticle": {Name: "AddParticle", ReturnType: "void", Parameters: []parameter{
			{Name: "pos", Type: "Vector3"}, {Name: "p", Type: "Particle"},
		}},
		"PlaySound": {Name: "PlaySound", ReturnType: "void", Parameters: []parameter{
			{Name: "pos", Type: "Vector3"}, {Name: "s", Type: "Sound"},
		}},
		"AddEntity": {Name: "AddEntity", ReturnType: "Entity", Parameters: []parameter{
			{Name: "e", Type: "EntityHandle"},
		}},
		"AddEntityAt": {Name: "AddEntityAt", ReturnType: "Entity", Parameters: []parameter{
			{Name: "e", Type: "EntityHandle"}, {Name: "pos", Type: "Vector3"},
		}},
		"RemoveEntity": {Name: "RemoveEntity", ReturnType: "EntityHandle", Parameters: []parameter{
			{Name: "e", Type: "Entity"},
		}},
		"Entities": {Name: "Entities", ReturnType: "IEnumerable<Entity>"},
		"EntitiesWithin": {Name: "EntitiesWithin", ReturnType: "IEnumerable<Entity>", Parameters: []parameter{
			{Name: "box", Type: "Cube.BBox"},
		}},
		"Players": {Name: "Players", ReturnType: "IEnumerable<Entity>"},
	}[method.Name]
	if !reflect.DeepEqual(method, expected) {
		return fmt.Errorf("signature changed: got %s %s(%s)", method.ReturnType, method.Name, formatParameters(method.Parameters))
	}
	return nil
}

func selectedWorldTxMethod(name string) bool {
	for _, selected := range selectedWorldTxMethods {
		if name == selected {
			return true
		}
	}
	return false
}

func pointerReceiver(function *ast.FuncDecl, name string) bool {
	if function.Recv == nil || len(function.Recv.List) != 1 {
		return false
	}
	pointer, ok := function.Recv.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	receiver, ok := pointer.X.(*ast.Ident)
	return ok && receiver.Name == name
}

func translateWorldTxParameters(method string, fields *ast.FieldList) ([]parameter, error) {
	var parameters []parameter
	if fields == nil {
		return parameters, nil
	}
	for _, field := range fields.List {
		if len(field.Names) == 0 {
			return nil, fmt.Errorf("unnamed parameter of type %s", formatGoExpression(field.Type))
		}
		for _, name := range field.Names {
			nullableInterface := method != "ScheduleBlockUpdate" || name.Name != "b"
			typeName, ok := worldTxCSharpType(field.Type, nullableInterface)
			if !ok {
				return nil, fmt.Errorf("unsupported parameter type %s", formatGoExpression(field.Type))
			}
			if (method == "AddEntity" || method == "AddEntityAt") && typeName == "EntityHandle?" {
				typeName = "EntityHandle"
			}
			parameters = append(parameters, parameter{Name: name.Name, Type: typeName})
		}
	}
	return parameters, nil
}

func translateWorldTxResult(method string, fields *ast.FieldList) (string, error) {
	if fields == nil || len(fields.List) == 0 {
		return "void", nil
	}
	var results []string
	for _, field := range fields.List {
		typeName, ok := worldTxCSharpType(field.Type, false)
		if !ok {
			return "", fmt.Errorf("unsupported return type %s", formatGoExpression(field.Type))
		}
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for range count {
			results = append(results, typeName)
		}
	}
	if len(results) == 1 {
		if method == "World" {
			if results[0] != "World?" {
				return "", fmt.Errorf("expected *World, got %s", results[0])
			}
			return "World", nil
		}
		if method == "RemoveEntity" && results[0] == "EntityHandle?" {
			return "EntityHandle", nil
		}
		return results[0], nil
	}
	if method == "BlockLoaded" || method == "Liquid" {
		valueType := map[string]string{"BlockLoaded": "Block", "Liquid": "Liquid"}[method]
		if !reflect.DeepEqual(results, []string{valueType, "bool"}) {
			return "", fmt.Errorf("expected (%s, bool), got (%s)", valueType, strings.Join(results, ", "))
		}
		return fmt.Sprintf("(%s? %s, bool Ok)", valueType, valueType), nil
	}
	return "(" + strings.Join(results, ", ") + ")", nil
}

func worldTxCSharpType(expression ast.Expr, parameter bool) (string, bool) {
	switch value := expression.(type) {
	case *ast.Ellipsis:
		typeName, ok := worldTxCSharpType(value.Elt, false)
		if !ok {
			return "", false
		}
		return "params " + typeName + "[]", true
	case *ast.StarExpr:
		typeName, ok := worldTxCSharpType(value.X, parameter)
		if !ok {
			return "", false
		}
		return strings.TrimSuffix(typeName, "?") + "?", true
	case *ast.Ident:
		typeName, ok := map[string]string{
			"Block":        "Block",
			"Biome":        "Biome",
			"Entity":       "Entity",
			"EntityHandle": "EntityHandle",
			"Liquid":       "Liquid",
			"Particle":     "Particle",
			"Sound":        "Sound",
			"SetOpts":      "SetOpts",
			"World":        "World",
			"bool":         "bool",
			"float64":      "double",
			"int":          "int",
			"int64":        "long",
			"uint8":        "byte",
		}[value.Name]
		if !ok {
			return "", false
		}
		if parameter && (value.Name == "Block" || value.Name == "Liquid") {
			return typeName + "?", true
		}
		return typeName, true
	case *ast.SelectorExpr:
		packageName, ok := value.X.(*ast.Ident)
		if !ok {
			return "", false
		}
		typeName, ok := map[string]string{
			"cube.BBox":     "Cube.BBox",
			"cube.Pos":      "Cube.Pos",
			"cube.Range":    "Cube.Range",
			"mgl64.Vec3":    "Vector3",
			"time.Duration": "TimeSpan",
		}[packageName.Name+"."+value.Sel.Name]
		return typeName, ok
	case *ast.IndexExpr:
		selector, ok := value.X.(*ast.SelectorExpr)
		if !ok {
			return "", false
		}
		packageName, ok := selector.X.(*ast.Ident)
		if !ok || packageName.Name != "iter" || selector.Sel.Name != "Seq" {
			return "", false
		}
		element, ok := worldTxCSharpType(value.Index, false)
		if !ok || element != "Cube.Pos" && element != "Entity" {
			return "", false
		}
		return "IEnumerable<" + element + ">", true
	default:
		return "", false
	}
}

func inspectBlocks(directory string) (blockSpec, error) {
	packages, err := parser.ParseDir(token.NewFileSet(), directory, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		return blockSpec{}, err
	}
	pkg, ok := packages["block"]
	if !ok {
		return blockSpec{}, fmt.Errorf("Dragonfly block package not found")
	}
	declarations := map[string]*ast.TypeSpec{}
	for _, file := range pkg.Files {
		for _, declaration := range file.Decls {
			gen, ok := declaration.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, raw := range gen.Specs {
				typeSpec, ok := raw.(*ast.TypeSpec)
				if ok {
					declarations[typeSpec.Name.Name] = typeSpec
				}
			}
		}
	}

	world.DefaultBlockRegistry.Finalize()
	registered := map[reflect.Type][]world.Block{}
	for _, value := range world.DefaultBlockRegistry.Blocks() {
		typeOf := reflect.TypeOf(value)
		if typeOf == nil {
			continue
		}
		if typeOf.Kind() == reflect.Pointer {
			typeOf = typeOf.Elem()
		}
		if typeOf.PkgPath() == "github.com/df-mc/dragonfly/server/block" && typeOf.Name() != "" {
			registered[typeOf] = append(registered[typeOf], value)
		}
	}

	var result blockSpec
	types := make([]reflect.Type, 0, len(registered))
	for typeOf := range registered {
		types = append(types, typeOf)
	}
	sort.Slice(types, func(i, j int) bool { return types[i].Name() < types[j].Name() })
	for _, typeOf := range types {
		if typeOf.Name() == "Water" || typeOf.Name() == "Lava" {
			continue
		}
		declaration := declarations[typeOf.Name()]
		if declaration == nil || !declaration.Name.IsExported() {
			return blockSpec{}, fmt.Errorf("registered block %s has no exported AST declaration", typeOf.Name())
		}
		visible := reflect.VisibleFields(typeOf)
		definition := blockTypeSpec{Name: typeOf.Name()}
		indexes := make([][]int, 0, len(visible))
		primitive := true
		for _, field := range visible {
			if !field.IsExported() {
				continue
			}
			fieldType := ""
			switch {
			case field.Type.PkgPath() == "" && field.Type.Name() == "bool":
				fieldType = "bool"
			case field.Type.PkgPath() == "" && field.Type.Name() == "int":
				fieldType = "int"
			default:
				primitive = false
			}
			if !primitive {
				break
			}
			if !blockFieldVaries(registered[typeOf], field.Index, fieldType) {
				continue
			}
			astField, err := blockFieldAtASTPath(typeOf.Name(), field.Index, declarations)
			if err != nil {
				return blockSpec{}, err
			}
			if astField.Name != field.Name || astField.Type != fieldType {
				return blockSpec{}, fmt.Errorf("Dragonfly block.%s promoted field %s differs between AST and reflection: got %#v", typeOf.Name(), field.Name, astField)
			}
			definition.Fields = append(definition.Fields, astField)
			indexes = append(indexes, field.Index)
		}
		if !primitive {
			continue
		}
		seen := map[string]encodedBlock{}
		for _, state := range registered[typeOf] {
			value := reflect.ValueOf(state)
			if value.Kind() == reflect.Pointer {
				value = value.Elem()
			}
			encoded, err := encodeRegisteredBlock(typeOf.Name(), state, registered[typeOf])
			if err != nil {
				return blockSpec{}, err
			}
			generated := encodedBlockState{encodedBlock: encoded}
			var key strings.Builder
			for index, path := range indexes {
				field := value.FieldByIndex(path)
				fieldValue := blockFieldValue{}
				switch definition.Fields[index].Type {
				case "bool":
					fieldValue.Bool = field.Bool()
					fmt.Fprintf(&key, "b:%t;", fieldValue.Bool)
				case "int":
					integer := field.Int()
					if integer < math.MinInt32 || integer > math.MaxInt32 {
						return blockSpec{}, fmt.Errorf("Dragonfly block.%s.%s value %d does not fit C# int", typeOf.Name(), definition.Fields[index].Name, integer)
					}
					fieldValue.Int = int(integer)
					fmt.Fprintf(&key, "i:%d;", fieldValue.Int)
				}
				generated.Values = append(generated.Values, fieldValue)
			}
			if previous, ok := seen[key.String()]; ok {
				if previous.Identifier != encoded.Identifier || !bytes.Equal(previous.PropertiesNBT, encoded.PropertiesNBT) {
					return blockSpec{}, fmt.Errorf("Dragonfly block.%s has duplicate public state %q with different encodings", typeOf.Name(), key.String())
				}
				continue
			}
			seen[key.String()] = encoded
			definition.States = append(definition.States, generated)
		}
		if len(definition.States) == 0 {
			return blockSpec{}, fmt.Errorf("Dragonfly block.%s has no registered states", typeOf.Name())
		}
		result.Types = append(result.Types, definition)
	}
	if len(result.Types) == 0 {
		return blockSpec{}, fmt.Errorf("no primitive Dragonfly block types found")
	}
	for _, name := range []string{"Water", "Lava"} {
		var liquidType reflect.Type
		var liquidStates []world.Block
		for typeOf, states := range registered {
			if typeOf.Name() == name {
				liquidType, liquidStates = typeOf, states
				break
			}
		}
		if liquidType == nil {
			return blockSpec{}, fmt.Errorf("Dragonfly block.%s is not registered", name)
		}
		if err := validateLiquidFields(declarations[name], name); err != nil {
			return blockSpec{}, err
		}
		liquid := liquidSpec{Name: name, States: make([]encodedLiquid, 0, len(liquidStates))}
		for _, state := range liquidStates {
			worldLiquid, ok := state.(world.Liquid)
			if !ok {
				return blockSpec{}, fmt.Errorf("registered block.%s state does not implement world.Liquid", name)
			}
			if liquid.LiquidType == "" {
				liquid.LiquidType = worldLiquid.LiquidType()
			} else if liquid.LiquidType != worldLiquid.LiquidType() {
				return blockSpec{}, fmt.Errorf("registered block.%s states have inconsistent liquid types", name)
			}
			value := reflect.ValueOf(state)
			if value.Kind() == reflect.Pointer {
				value = value.Elem()
			}
			encoded, err := encodeRegisteredBlock(name, state, liquidStates)
			if err != nil {
				return blockSpec{}, err
			}
			liquid.States = append(liquid.States, encodedLiquid{
				encodedBlock: encoded,
				Still:        value.FieldByName("Still").Bool(),
				Depth:        int(value.FieldByName("Depth").Int()),
				Falling:      value.FieldByName("Falling").Bool(),
			})
		}
		if len(liquid.States) == 0 {
			return blockSpec{}, fmt.Errorf("Dragonfly block.%s has no registered states", name)
		}
		result.Liquids = append(result.Liquids, liquid)
	}
	return result, nil
}

func inspectBiomes(directory string) ([]encodedBiome, error) {
	packages, err := parser.ParseDir(token.NewFileSet(), directory, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, err
	}
	pkg, ok := packages["biome"]
	if !ok {
		return nil, fmt.Errorf("Dragonfly biome package not found")
	}
	declarations := map[string]*ast.TypeSpec{}
	for _, file := range pkg.Files {
		for _, declaration := range file.Decls {
			gen, ok := declaration.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, raw := range gen.Specs {
				typeSpec, ok := raw.(*ast.TypeSpec)
				if ok && typeSpec.Name.IsExported() {
					declarations[typeSpec.Name.Name] = typeSpec
				}
			}
		}
	}

	registered := world.Biomes()
	if len(registered) == 0 {
		return nil, fmt.Errorf("Dragonfly has no registered biomes")
	}
	result := make([]encodedBiome, 0, len(registered))
	ids := map[int]string{}
	for _, value := range registered {
		typeOf := reflect.TypeOf(value)
		if typeOf == nil {
			return nil, fmt.Errorf("Dragonfly registered a nil biome")
		}
		if typeOf.Kind() == reflect.Pointer {
			typeOf = typeOf.Elem()
		}
		if typeOf.PkgPath() != "github.com/df-mc/dragonfly/server/world/biome" || !ast.IsExported(typeOf.Name()) {
			return nil, fmt.Errorf("registered biome %s is not a vanilla biome type", typeOf)
		}
		spec := declarations[typeOf.Name()]
		structure, ok := biomeEmptyStruct(spec)
		if !ok || len(structure.Fields.List) != 0 {
			return nil, fmt.Errorf("Dragonfly biome.%s is not an empty struct", typeOf.Name())
		}
		id := value.EncodeBiome()
		if id < -1<<31 || id > 1<<31-1 {
			return nil, fmt.Errorf("Dragonfly biome.%s ID %d does not fit C# int", typeOf.Name(), id)
		}
		if previous, exists := ids[id]; exists {
			return nil, fmt.Errorf("Dragonfly biomes %s and %s share ID %d", previous, typeOf.Name(), id)
		}
		ids[id] = typeOf.Name()
		result = append(result, encodedBiome{Name: typeOf.Name(), ID: id})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func biomeEmptyStruct(spec *ast.TypeSpec) (*ast.StructType, bool) {
	if spec == nil {
		return nil, false
	}
	structure, ok := spec.Type.(*ast.StructType)
	return structure, ok
}

func inspectItems(directory string, effects effectSpec) (itemSpec, error) {
	packages, err := parser.ParseDir(token.NewFileSet(), directory, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		return itemSpec{}, err
	}
	pkg, ok := packages["item"]
	if !ok {
		return itemSpec{}, fmt.Errorf("Dragonfly item package not found")
	}
	types := map[string]*ast.TypeSpec{}
	variables := map[string]bool{}
	functions := map[string]*ast.FuncDecl{}
	methods := map[string]map[string]*ast.FuncDecl{}
	for _, file := range pkg.Files {
		for _, declaration := range file.Decls {
			switch value := declaration.(type) {
			case *ast.GenDecl:
				for _, raw := range value.Specs {
					switch spec := raw.(type) {
					case *ast.TypeSpec:
						types[spec.Name.Name] = spec
					case *ast.ValueSpec:
						for _, name := range spec.Names {
							variables[name.Name] = true
						}
					}
				}
			case *ast.FuncDecl:
				if value.Recv == nil {
					functions[value.Name.Name] = value
					continue
				}
				receiver, ok := receiverName(value)
				if !ok {
					continue
				}
				if methods[receiver] == nil {
					methods[receiver] = map[string]*ast.FuncDecl{}
				}
				methods[receiver][value.Name.Name] = value
			}
		}
	}

	tierFields, err := inspectToolTierFields(types["ToolTier"])
	if err != nil {
		return itemSpec{}, err
	}
	tierNames, err := inspectToolTierVariables(functions["ToolTiers"], variables)
	if err != nil {
		return itemSpec{}, err
	}
	liveTiers := dfitem.ToolTiers()
	if len(tierNames) != len(liveTiers) {
		return itemSpec{}, fmt.Errorf("Dragonfly item.ToolTiers AST/live lengths differ: %d/%d", len(tierNames), len(liveTiers))
	}
	valueTypes, err := inspectItemValueTypes(directory, functions, effects)
	if err != nil {
		return itemSpec{}, err
	}
	enchantments, err := inspectEnchantments(filepath.Join(directory, "enchantment"))
	if err != nil {
		return itemSpec{}, err
	}
	result := itemSpec{
		ToolTierFields: tierFields,
		ToolTiers:      make([]toolTierSpec, len(liveTiers)),
		ValueTypes:     valueTypes,
		Enchantments:   enchantments,
	}
	for index, tier := range liveTiers {
		result.ToolTiers[index] = toolTierSpec{Variable: tierNames[index], Value: tier}
	}

	registered := map[reflect.Type][]world.Item{}
	for _, value := range world.Items() {
		typeOf := reflect.TypeOf(value)
		if typeOf == nil {
			continue
		}
		if typeOf.Kind() == reflect.Pointer {
			typeOf = typeOf.Elem()
		}
		if typeOf.PkgPath() == "github.com/df-mc/dragonfly/server/block" && typeOf.Name() == "Air" {
			result.AirIdentifier, _ = value.EncodeItem()
		}
		if typeOf.Kind() == reflect.Struct && typeOf.PkgPath() == "github.com/df-mc/dragonfly/server/item" && ast.IsExported(typeOf.Name()) {
			registered[typeOf] = append(registered[typeOf], value)
		}
	}
	if result.AirIdentifier == "" {
		return itemSpec{}, fmt.Errorf("Dragonfly registered block.Air item not found")
	}
	armour, armourTypes, err := inspectArmourItems(types, functions, methods, registered, liveTiers, valueTypes)
	if err != nil {
		return itemSpec{}, err
	}
	result.Armour = armour
	result.Types = append(result.Types, armourTypes...)

	for typeOf, values := range registered {
		declaration := types[typeOf.Name()]
		if itemTypeByName(armourTypes, typeOf.Name()) != nil {
			continue
		}
		if typeOf.Name() == "Bucket" {
			definition, bucket, err := inspectBucketItem(typeOf, values, declaration, types, functions, methods, liveTiers, valueTypes)
			if err != nil {
				return itemSpec{}, err
			}
			result.Bucket = bucket
			result.Types = append(result.Types, definition)
			continue
		}
		if typeOf.Name() == "Crossbow" {
			definition, crossbow, err := inspectCrossbowItem(typeOf, values, declaration, types, methods, liveTiers, valueTypes)
			if err != nil {
				return itemSpec{}, err
			}
			result.Crossbow = crossbow
			result.Types = append(result.Types, definition)
			continue
		}
		if typeOf.Name() == "BookAndQuill" || typeOf.Name() == "WrittenBook" {
			if len(values) != 1 {
				return itemSpec{}, fmt.Errorf("Dragonfly item.%s registry states changed: %d", typeOf.Name(), len(values))
			}
			if err := validateBookItemType(declaration, typeOf.Name(), methods[typeOf.Name()]); err != nil {
				return itemSpec{}, err
			}
			definition := itemTypeSpec{Name: typeOf.Name(), NBT: true}
			state, err := inspectItemState(typeOf, values[0], nil, liveTiers, valueTypes)
			if err != nil {
				return itemSpec{}, err
			}
			definition.States = []itemStateSpec{state}
			result.Types = append(result.Types, definition)
			continue
		}
		if typeOf.Name() == "Firework" || typeOf.Name() == "FireworkStar" {
			definition, err := inspectFireworkItemType(typeOf, values, declaration, types, methods, liveTiers, valueTypes)
			if err != nil {
				return itemSpec{}, err
			}
			result.Types = append(result.Types, definition)
			continue
		}
		fields, supported, err := inspectItemFields(declaration, typeOf.Name(), valueTypes)
		if err != nil {
			return itemSpec{}, err
		}
		if !supported {
			continue
		}
		definition := itemTypeSpec{Name: typeOf.Name(), Fields: fields}
		states := values
		if len(fields) == 1 && fields[0].Kind == itemFieldValue {
			valueType := findItemValueType(valueTypes, fields[0].ValueType)
			states = make([]world.Item, 0, len(valueType.Values))
			for _, fieldValue := range valueType.Values {
				reflected := reflect.New(typeOf).Elem()
				reflected.FieldByName(fields[0].Name).Set(reflect.ValueOf(fieldValue))
				states = append(states, reflected.Interface().(world.Item))
			}
		}
		for _, value := range states {
			state, err := inspectItemState(typeOf, value, fields, liveTiers, valueTypes)
			if err != nil {
				return itemSpec{}, err
			}
			definition.States = append(definition.States, state)
		}
		if !completeItemType(typeOf, definition, liveTiers, valueTypes) {
			continue
		}
		sort.Slice(definition.States, func(i, j int) bool {
			left, right := definition.States[i], definition.States[j]
			if left.ToolTier != right.ToolTier {
				return left.ToolTier < right.ToolTier
			}
			for index := range left.Values {
				if left.Values[index] != right.Values[index] {
					return left.Values[index] < right.Values[index]
				}
			}
			for index := range left.Bools {
				if left.Bools[index] != right.Bools[index] {
					return !left.Bools[index]
				}
			}
			if left.Identifier != right.Identifier {
				return left.Identifier < right.Identifier
			}
			return left.Metadata < right.Metadata
		})
		result.Types = append(result.Types, definition)
	}
	sort.Slice(result.Types, func(i, j int) bool { return result.Types[i].Name < result.Types[j].Name })
	if len(result.Types) == 0 {
		return itemSpec{}, fmt.Errorf("no safely representable Dragonfly items found")
	}
	return result, nil
}

func inspectEnchantments(directory string) ([]enchantmentSpec, error) {
	packages, err := parser.ParseDir(token.NewFileSet(), directory, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, err
	}
	pkg, ok := packages["enchantment"]
	if !ok {
		return nil, fmt.Errorf("Dragonfly item/enchantment package not found")
	}
	exportedVariables := map[string]bool{}
	registeredNames := map[int]string{}
	for _, file := range pkg.Files {
		for _, declaration := range file.Decls {
			switch value := declaration.(type) {
			case *ast.GenDecl:
				if value.Tok != token.VAR {
					continue
				}
				for _, raw := range value.Specs {
					spec, ok := raw.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for _, name := range spec.Names {
						if ast.IsExported(name.Name) {
							exportedVariables[name.Name] = true
						}
					}
				}
			case *ast.FuncDecl:
				if value.Recv != nil || value.Name.Name != "init" || value.Body == nil {
					continue
				}
				ast.Inspect(value.Body, func(node ast.Node) bool {
					call, ok := node.(*ast.CallExpr)
					if !ok || len(call.Args) != 2 {
						return true
					}
					selector, ok := call.Fun.(*ast.SelectorExpr)
					if !ok || selector.Sel.Name != "RegisterEnchantment" {
						return true
					}
					idLiteral, idOK := call.Args[0].(*ast.BasicLit)
					name, nameOK := call.Args[1].(*ast.Ident)
					if !idOK || idLiteral.Kind != token.INT || !nameOK {
						return true
					}
					id, parseErr := strconv.Atoi(idLiteral.Value)
					if parseErr == nil {
						registeredNames[id] = name.Name
					}
					return true
				})
			}
		}
	}
	live := dfitem.Enchantments()
	if len(live) != len(registeredNames) {
		return nil, fmt.Errorf("Dragonfly enchantment AST/live lengths differ: %d/%d", len(registeredNames), len(live))
	}
	result := make([]enchantmentSpec, 0, len(live))
	for _, enchantment := range live {
		id, found := dfitem.EnchantmentID(enchantment)
		if !found || id < 0 || id >= 64 {
			return nil, fmt.Errorf("Dragonfly enchantment %q has unsupported ID %d", enchantment.Name(), id)
		}
		name, found := registeredNames[id]
		if !found || !exportedVariables[name] {
			return nil, fmt.Errorf("Dragonfly enchantment ID %d has no exported AST variable", id)
		}
		entry := enchantmentSpec{Name: name, ID: id, DisplayName: enchantment.Name(), MaxLevel: enchantment.MaxLevel()}
		for _, other := range live {
			otherID, found := dfitem.EnchantmentID(other)
			if !found || otherID < 0 || otherID >= 64 {
				return nil, fmt.Errorf("Dragonfly enchantment %q has unsupported ID %d", other.Name(), otherID)
			}
			if enchantment.CompatibleWithEnchantment(other) {
				entry.CompatibleEnchantments |= uint64(1) << otherID
			}
		}
		seenItems := map[encodedItemKey]bool{}
		for _, value := range world.Items() {
			if !enchantment.CompatibleWithItem(value) {
				continue
			}
			identifier, metadata := value.EncodeItem()
			key := encodedItemKey{Identifier: identifier, Metadata: int(metadata)}
			if !seenItems[key] {
				seenItems[key] = true
				entry.CompatibleItems = append(entry.CompatibleItems, key)
			}
		}
		sort.Slice(entry.CompatibleItems, func(i, j int) bool {
			left, right := entry.CompatibleItems[i], entry.CompatibleItems[j]
			if left.Identifier != right.Identifier {
				return left.Identifier < right.Identifier
			}
			return left.Metadata < right.Metadata
		})
		result = append(result, entry)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func inspectBucketItem(
	typeOf reflect.Type,
	values []world.Item,
	declaration *ast.TypeSpec,
	types map[string]*ast.TypeSpec,
	functions map[string]*ast.FuncDecl,
	methods map[string]map[string]*ast.FuncDecl,
	toolTiers []dfitem.ToolTier,
	valueTypes []itemValueTypeSpec,
) (itemTypeSpec, bucketSpec, error) {
	if fields := exportedItemFields(declaration); !reflect.DeepEqual(fields, []string{"Content BucketContent"}) {
		return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.Bucket fields changed: %v", fields)
	}
	if fields := allItemFields(types["BucketContent"]); !reflect.DeepEqual(fields, []string{"liquid world.Liquid", "milk bool"}) {
		return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.BucketContent fields changed: %v", fields)
	}
	for name, signature := range map[string]goSignature{
		"LiquidBucketContent": {Parameters: "world.Liquid", Results: "BucketContent"},
		"MilkBucketContent":   {Results: "BucketContent"},
	} {
		function := functions[name]
		if function == nil || function.Recv != nil || rawParameterTypes(function.Type.Params) != signature.Parameters || rawResultTypes(function.Type.Results) != signature.Results {
			return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.%s signature changed", name)
		}
	}
	for name, signature := range map[string]goSignature{
		"Liquid":     {Results: "world.Liquid,bool"},
		"String":     {Results: "string"},
		"LiquidType": {Results: "string"},
	} {
		method := methods["BucketContent"][name]
		if method == nil || !valueReceiver(method, "BucketContent") || rawParameterTypes(method.Type.Params) != signature.Parameters || rawResultTypes(method.Type.Results) != signature.Results {
			return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.BucketContent.%s signature changed", name)
		}
	}
	for name, signature := range map[string]goSignature{
		"MaxCount":         {Results: "int"},
		"AlwaysConsumable": {Results: "bool"},
		"CanConsume":       {Results: "bool"},
		"ConsumeDuration":  {Results: "time.Duration"},
		"Consume":          {Parameters: "*world.Tx,Consumer", Results: "Stack"},
		"Empty":            {Results: "bool"},
		"FuelInfo":         {Results: "FuelInfo"},
		"UseOnBlock":       {Parameters: "cube.Pos,cube.Face,mgl64.Vec3,*world.Tx,User,*UseContext", Results: "bool"},
		"EncodeItem":       {Results: "string,int16"},
	} {
		method := methods["Bucket"][name]
		if method == nil || !valueReceiver(method, "Bucket") || rawParameterTypes(method.Type.Params) != signature.Parameters || rawResultTypes(method.Type.Results) != signature.Results {
			return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.Bucket.%s signature changed", name)
		}
	}
	for name, wanted := range map[string][]string{
		"MaxCounter":    {"MaxCount()->int"},
		"Fuel":          {"FuelInfo()->FuelInfo"},
		"Consumable":    {"AlwaysConsumable()->bool", "ConsumeDuration()->time.Duration", "Consume(*world.Tx,Consumer)->Stack"},
		"UsableOnBlock": {"UseOnBlock(cube.Pos,cube.Face,mgl64.Vec3,*world.Tx,User,*UseContext)->bool"},
	} {
		if err := validateItemInterface(types[name], name, wanted); err != nil {
			return itemTypeSpec{}, bucketSpec{}, err
		}
	}

	expected := []struct {
		identifier string
		kind       bucketContentKind
	}{
		{"minecraft:bucket", bucketEmpty},
		{"minecraft:water_bucket", bucketWater},
		{"minecraft:lava_bucket", bucketLava},
		{"minecraft:milk_bucket", bucketMilk},
	}
	if len(values) != len(expected) {
		return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.Bucket registry states changed: %d", len(values))
	}
	byIdentifier := make(map[string]dfitem.Bucket, len(values))
	for _, raw := range values {
		bucket, ok := raw.(dfitem.Bucket)
		if !ok {
			return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly registered Bucket has type %T", raw)
		}
		identifier, metadata := bucket.EncodeItem()
		if _, exists := byIdentifier[identifier]; metadata != 0 || exists {
			return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.Bucket registered duplicate/metadata state %s:%d", identifier, metadata)
		}
		byIdentifier[identifier] = bucket
	}
	definition := itemTypeSpec{Name: "Bucket", Bucket: true}
	result := bucketSpec{Present: true, ConsumeDuration: dfitem.DefaultConsumeDuration}
	for _, wanted := range expected {
		bucket, ok := byIdentifier[wanted.identifier]
		if !ok {
			return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.Bucket state %s missing", wanted.identifier)
		}
		if _, ok := any(bucket).(dfitem.MaxCounter); !ok {
			return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.Bucket no longer implements MaxCounter")
		}
		if _, ok := any(bucket).(dfitem.Fuel); !ok {
			return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.Bucket no longer implements Fuel")
		}
		if _, ok := any(bucket).(dfitem.Consumable); !ok {
			return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.Bucket no longer implements Consumable")
		}
		if _, ok := any(bucket).(dfitem.UsableOnBlock); !ok {
			return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.Bucket no longer implements UsableOnBlock")
		}
		state, err := inspectItemState(typeOf, bucket, nil, toolTiers, valueTypes)
		if err != nil {
			return itemTypeSpec{}, bucketSpec{}, err
		}
		state.Bucket = wanted.kind
		liquid, liquidOK := bucket.Content.Liquid()
		liquidName := ""
		if liquidOK {
			liquidType := reflect.TypeOf(liquid)
			if liquidType.Kind() == reflect.Pointer {
				liquidType = liquidType.Elem()
			}
			liquidName = liquidType.PkgPath() + "." + liquidType.Name()
		}
		wantLiquid := map[bucketContentKind]string{
			bucketWater: "github.com/df-mc/dragonfly/server/block.Water",
			bucketLava:  "github.com/df-mc/dragonfly/server/block.Lava",
		}[wanted.kind]
		wantEmpty, wantMilk := wanted.kind == bucketEmpty, wanted.kind == bucketMilk
		if bucket.Empty() != wantEmpty || bucket.AlwaysConsumable() != wantMilk || bucket.CanConsume() != wantMilk ||
			bucket.ConsumeDuration() != dfitem.DefaultConsumeDuration || liquidOK != (wantLiquid != "") || liquidName != wantLiquid {
			return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.Bucket state %s behaviour changed", wanted.identifier)
		}
		wantString := map[bucketContentKind]string{bucketEmpty: "", bucketWater: "water", bucketLava: "lava", bucketMilk: "milk"}[wanted.kind]
		wantLiquidType := wantString
		if wanted.kind == bucketEmpty {
			wantLiquidType = "milk"
		}
		if bucket.Content.String() != wantString || bucket.Content.LiquidType() != wantLiquidType {
			return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.BucketContent state %s behaviour changed", wanted.identifier)
		}
		if liquidOK {
			created := dfitem.LiquidBucketContent(liquid)
			createdLiquid, createdOK := created.Liquid()
			if !createdOK || !reflect.DeepEqual(createdLiquid, liquid) || created.String() != wantString || created.LiquidType() != wantLiquidType {
				return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.LiquidBucketContent state %s behaviour changed", wanted.identifier)
			}
		} else if wanted.kind == bucketMilk {
			created := dfitem.MilkBucketContent()
			if _, ok := created.Liquid(); ok || created.String() != "milk" || created.LiquidType() != "milk" {
				return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.MilkBucketContent behaviour changed")
			}
		}
		if wanted.kind == bucketLava {
			if state.Capability.FuelDuration != 1000*time.Second || state.Capability.FuelIdentifier != "minecraft:bucket" || state.Capability.FuelMetadata != 0 || state.Capability.FuelCount != 1 {
				return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.Bucket lava fuel changed: %#v", state.Capability)
			}
			result.FuelDuration = state.Capability.FuelDuration
			result.FuelResidueIdentifier = state.Capability.FuelIdentifier
			result.FuelResidueMetadata = state.Capability.FuelMetadata
			result.FuelResidueCount = state.Capability.FuelCount
		} else if state.Capability.FuelDuration != 0 || state.Capability.FuelIdentifier != "" || state.Capability.FuelCount != 0 {
			return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly item.Bucket state %s fuel changed: %#v", wanted.identifier, state.Capability)
		}
		if wanted.kind == bucketEmpty {
			result.EmptyMaxCount = state.Capability.MaxCount
		} else if result.FullMaxCount == 0 {
			result.FullMaxCount = state.Capability.MaxCount
		} else if result.FullMaxCount != state.Capability.MaxCount {
			return itemTypeSpec{}, bucketSpec{}, fmt.Errorf("Dragonfly filled Bucket max counts differ")
		}
		definition.States = append(definition.States, state)
	}
	return definition, result, nil
}

func inspectCrossbowItem(
	typeOf reflect.Type,
	values []world.Item,
	declaration *ast.TypeSpec,
	types map[string]*ast.TypeSpec,
	methods map[string]map[string]*ast.FuncDecl,
	toolTiers []dfitem.ToolTier,
	valueTypes []itemValueTypeSpec,
) (itemTypeSpec, crossbowSpec, error) {
	if len(values) != 1 {
		return itemTypeSpec{}, crossbowSpec{}, fmt.Errorf("Dragonfly item.Crossbow registry states changed: %d", len(values))
	}
	if fields := exportedItemFields(declaration); !reflect.DeepEqual(fields, []string{"Item Stack"}) {
		return itemTypeSpec{}, crossbowSpec{}, fmt.Errorf("Dragonfly item.Crossbow fields changed: %v", fields)
	}
	for name, signature := range map[string]goSignature{
		"MaxCount":         {Results: "int"},
		"DurabilityInfo":   {Results: "DurabilityInfo"},
		"FuelInfo":         {Results: "FuelInfo"},
		"EnchantmentValue": {Results: "int"},
		"Charge":           {Parameters: "Releaser,*world.Tx,*UseContext,time.Duration", Results: "bool"},
		"ContinueCharge":   {Parameters: "Releaser,*world.Tx,*UseContext,time.Duration"},
		"ReleaseCharge":    {Parameters: "Releaser,*world.Tx,*UseContext", Results: "bool"},
		"CanCharge":        {Parameters: "Releaser,*world.Tx,*UseContext", Results: "bool"},
		"EncodeItem":       {Results: "string,int16"},
		"DecodeNBT":        {Parameters: "map[string]any", Results: "any"},
		"EncodeNBT":        {Results: "map[string]any"},
	} {
		method := methods["Crossbow"][name]
		if method == nil || !valueReceiver(method, "Crossbow") || rawParameterTypes(method.Type.Params) != signature.Parameters || rawResultTypes(method.Type.Results) != signature.Results {
			return itemTypeSpec{}, crossbowSpec{}, fmt.Errorf("Dragonfly item.Crossbow.%s signature changed", name)
		}
	}
	if err := validateItemInterface(types["Fuel"], "Fuel", []string{"FuelInfo()->FuelInfo"}); err != nil {
		return itemTypeSpec{}, crossbowSpec{}, err
	}
	if err := validateItemInterface(types["Chargeable"], "Chargeable", []string{
		"Charge(Releaser,*world.Tx,*UseContext,time.Duration)->bool",
		"ContinueCharge(Releaser,*world.Tx,*UseContext,time.Duration)->",
		"ReleaseCharge(Releaser,*world.Tx,*UseContext)->bool",
		"CanCharge(Releaser,*world.Tx,*UseContext)->bool",
	}); err != nil {
		return itemTypeSpec{}, crossbowSpec{}, err
	}
	if fields := exportedItemFields(types["FuelInfo"]); !reflect.DeepEqual(fields, []string{"Duration time.Duration", "Residue Stack"}) {
		return itemTypeSpec{}, crossbowSpec{}, fmt.Errorf("Dragonfly item.FuelInfo fields changed: %v", fields)
	}
	withResidue := methods["FuelInfo"]["WithResidue"]
	if withResidue == nil || !valueReceiver(withResidue, "FuelInfo") || rawParameterTypes(withResidue.Type.Params) != "Stack" || rawResultTypes(withResidue.Type.Results) != "FuelInfo" {
		return itemTypeSpec{}, crossbowSpec{}, fmt.Errorf("Dragonfly item.FuelInfo.WithResidue signature changed")
	}
	value, ok := values[0].(dfitem.Crossbow)
	if !ok {
		return itemTypeSpec{}, crossbowSpec{}, fmt.Errorf("Dragonfly registered Crossbow has type %T", values[0])
	}
	if _, ok := any(value).(dfitem.Chargeable); !ok {
		return itemTypeSpec{}, crossbowSpec{}, fmt.Errorf("Dragonfly item.Crossbow no longer implements Chargeable")
	}
	state, err := inspectItemState(typeOf, value, nil, toolTiers, valueTypes)
	if err != nil {
		return itemTypeSpec{}, crossbowSpec{}, err
	}
	durability := value.DurabilityInfo()
	fuel := value.FuelInfo()
	if !durability.BrokenItem().Empty() || durability.Persistent || durability.AttackDurability != 0 || durability.BreakDurability != 0 || !fuel.Residue.Empty() {
		return itemTypeSpec{}, crossbowSpec{}, fmt.Errorf("Dragonfly item.Crossbow capability shape changed")
	}
	return itemTypeSpec{Name: "Crossbow", NBT: true, States: []itemStateSpec{state}}, crossbowSpec{
		Present: true, MaxCount: value.MaxCount(), MaxDurability: durability.MaxDurability,
		EnchantmentValue: value.EnchantmentValue(), FuelDuration: fuel.Duration,
	}, nil
}

func inspectArmourItems(
	types map[string]*ast.TypeSpec,
	functions map[string]*ast.FuncDecl,
	methods map[string]map[string]*ast.FuncDecl,
	registered map[reflect.Type][]world.Item,
	toolTiers []dfitem.ToolTier,
	valueTypes []itemValueTypeSpec,
) (armourSpec, []itemTypeSpec, error) {
	for name, wanted := range map[string][]string{
		"Armour":             {"DefencePoints()->float64", "Toughness()->float64", "KnockBackResistance()->float64"},
		"ArmourTier":         {"BaseDurability()->float64", "Toughness()->float64", "KnockBackResistance()->float64", "EnchantmentValue()->int", "Name()->string"},
		"HelmetType":         {"embed:Armour", "Helmet()->bool"},
		"ChestplateType":     {"embed:Armour", "Chestplate()->bool"},
		"LeggingsType":       {"embed:Armour", "Leggings()->bool"},
		"BootsType":          {"embed:Armour", "Boots()->bool"},
		"ArmourTrimMaterial": {"TrimMaterial()->string", "MaterialColour()->string"},
		"Trimmable":          {"WithTrim(ArmourTrim)->world.Item"},
		"MaxCounter":         {"MaxCount()->int"},
		"Enchantable":        {"EnchantmentValue()->int"},
		"Durable":            {"DurabilityInfo()->DurabilityInfo"},
		"Repairable":         {"embed:Durable", "RepairableBy(Stack)->bool"},
		"Smeltable":          {"SmeltInfo()->SmeltInfo"},
	} {
		if err := validateItemInterface(types[name], name, wanted); err != nil {
			return armourSpec{}, nil, err
		}
	}
	if fields := exportedItemFields(types["ArmourTrim"]); !reflect.DeepEqual(fields, []string{"Template SmithingTemplateType", "Material ArmourTrimMaterial"}) {
		return armourSpec{}, nil, fmt.Errorf("Dragonfly item.ArmourTrim fields changed: %v", fields)
	}
	if fields := exportedItemFields(types["DurabilityInfo"]); !reflect.DeepEqual(fields, []string{
		"MaxDurability int", "BrokenItem func() Stack", "AttackDurability int", "BreakDurability int", "Persistent bool",
	}) {
		return armourSpec{}, nil, fmt.Errorf("Dragonfly item.DurabilityInfo fields changed: %v", fields)
	}
	if fields := exportedItemFields(types["SmeltInfo"]); !reflect.DeepEqual(fields, []string{
		"Product Stack", "Experience float64", "Food bool", "Ores bool",
	}) {
		return armourSpec{}, nil, fmt.Errorf("Dragonfly item.SmeltInfo fields changed: %v", fields)
	}
	zeroMethod := methods["ArmourTrim"]["Zero"]
	if zeroMethod == nil || !valueReceiver(zeroMethod, "ArmourTrim") || rawParameterTypes(zeroMethod.Type.Params) != "" || rawResultTypes(zeroMethod.Type.Results) != "bool" {
		return armourSpec{}, nil, fmt.Errorf("Dragonfly item.ArmourTrim.Zero signature changed")
	}
	liveTiers := dfitem.ArmourTiers()
	tierNames, err := inspectArmourTierNames(functions["ArmourTiers"])
	if err != nil {
		return armourSpec{}, nil, err
	}
	if len(tierNames) != len(liveTiers) {
		return armourSpec{}, nil, fmt.Errorf("Dragonfly item.ArmourTiers AST/live lengths differ: %d/%d", len(tierNames), len(liveTiers))
	}
	result := armourSpec{Tiers: make([]armourTierSpec, len(liveTiers))}
	for index, tier := range liveTiers {
		typeOf := reflect.TypeOf(tier)
		if typeOf.Name() != tierNames[index] {
			return armourSpec{}, nil, fmt.Errorf("Dragonfly item.ArmourTiers AST/live order differs at %d: %s/%s", index, tierNames[index], typeOf.Name())
		}
		if err := validateArmourTierType(types[typeOf.Name()], typeOf, methods[typeOf.Name()]); err != nil {
			return armourSpec{}, nil, err
		}
		result.Tiers[index] = armourTierSpec{
			Name:                typeOf.Name(),
			BaseDurability:      tier.BaseDurability(),
			Toughness:           tier.Toughness(),
			KnockBackResistance: tier.KnockBackResistance(),
			EnchantmentValue:    tier.EnchantmentValue(),
			IdentifierName:      tier.Name(),
			Colour:              typeOf.NumField() == 1,
		}
	}

	var definitions []itemTypeSpec
	for typeOf, values := range registered {
		if len(values) == 0 {
			continue
		}
		if _, ok := values[0].(dfitem.Armour); !ok {
			continue
		}
		if !reflect.DeepEqual(exportedItemFields(types[typeOf.Name()]), []string{"Tier ArmourTier", "Trim ArmourTrim"}) {
			continue
		}
		piece, definition, err := inspectArmourPiece(typeOf, values, types[typeOf.Name()], methods[typeOf.Name()], liveTiers, toolTiers, valueTypes)
		if err != nil {
			return armourSpec{}, nil, err
		}
		result.Pieces = append(result.Pieces, piece)
		definitions = append(definitions, definition)
	}
	if len(result.Pieces) != 4 {
		return armourSpec{}, nil, fmt.Errorf("Dragonfly armour piece count changed: %d", len(result.Pieces))
	}
	sort.Slice(result.Pieces, func(i, j int) bool { return result.Pieces[i].Name < result.Pieces[j].Name })
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Name < definitions[j].Name })

	for _, value := range dfitem.ArmourTrimMaterials() {
		material, ok := value.(dfitem.ArmourTrimMaterial)
		if !ok {
			return armourSpec{}, nil, fmt.Errorf("Dragonfly armour trim material %T does not implement item.ArmourTrimMaterial", value)
		}
		typeOf := reflect.TypeOf(value)
		if typeOf.Kind() == reflect.Pointer {
			typeOf = typeOf.Elem()
		}
		if typeOf.PkgPath() != "github.com/df-mc/dragonfly/server/item" || typeOf.Name() == "" || types[typeOf.Name()] == nil {
			return armourSpec{}, nil, fmt.Errorf("Dragonfly armour trim material %T is unsupported", value)
		}
		for methodName := range map[string]bool{"TrimMaterial": true, "MaterialColour": true} {
			method := methods[typeOf.Name()][methodName]
			if method == nil || !valueReceiver(method, typeOf.Name()) || rawParameterTypes(method.Type.Params) != "" || rawResultTypes(method.Type.Results) != "string" {
				return armourSpec{}, nil, fmt.Errorf("Dragonfly item.%s.%s signature changed", typeOf.Name(), methodName)
			}
		}
		result.TrimMaterials = append(result.TrimMaterials, armourTrimMaterialSpec{
			ItemName: typeOf.Name(), Material: material.TrimMaterial(), MaterialColour: material.MaterialColour(),
		})
		if _, registeredDirectly := registered[typeOf]; !registeredDirectly {
			if fields := exportedItemFields(types[typeOf.Name()]); len(fields) != 0 {
				return armourSpec{}, nil, fmt.Errorf("Dragonfly unregistered armour trim material %s has fields: %v", typeOf.Name(), fields)
			}
			state, err := inspectItemState(typeOf, value, nil, toolTiers, valueTypes)
			if err != nil {
				return armourSpec{}, nil, err
			}
			definitions = append(definitions, itemTypeSpec{Name: typeOf.Name(), States: []itemStateSpec{state}})
		}
	}
	if len(result.TrimMaterials) != 11 {
		return armourSpec{}, nil, fmt.Errorf("Dragonfly armour trim material count changed: %d", len(result.TrimMaterials))
	}
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Name < definitions[j].Name })
	return result, definitions, nil
}

func validateItemInterface(spec *ast.TypeSpec, name string, wanted []string) error {
	if spec == nil {
		return fmt.Errorf("Dragonfly item.%s interface missing", name)
	}
	interfaceType, ok := spec.Type.(*ast.InterfaceType)
	if !ok {
		return fmt.Errorf("Dragonfly item.%s is not an interface", name)
	}
	var members []string
	for _, field := range interfaceType.Methods.List {
		if len(field.Names) == 0 {
			members = append(members, "embed:"+formatGoExpression(field.Type))
			continue
		}
		method, ok := field.Type.(*ast.FuncType)
		if !ok || len(field.Names) != 1 {
			return fmt.Errorf("Dragonfly item.%s contains unsupported member", name)
		}
		members = append(members, field.Names[0].Name+"("+rawParameterTypes(method.Params)+")->"+rawResultTypes(method.Results))
	}
	if !reflect.DeepEqual(members, wanted) {
		return fmt.Errorf("Dragonfly item.%s interface changed: %v", name, members)
	}
	return nil
}

func inspectArmourTierNames(function *ast.FuncDecl) ([]string, error) {
	if function == nil || function.Type.Params == nil || function.Type.Params.NumFields() != 0 ||
		function.Type.Results == nil || len(function.Type.Results.List) != 1 ||
		formatGoExpression(function.Type.Results.List[0].Type) != "[]ArmourTier" || function.Body == nil || len(function.Body.List) != 1 {
		return nil, fmt.Errorf("Dragonfly item.ArmourTiers signature/body changed")
	}
	statement, ok := function.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(statement.Results) != 1 {
		return nil, fmt.Errorf("Dragonfly item.ArmourTiers body changed")
	}
	literal, ok := statement.Results[0].(*ast.CompositeLit)
	if !ok || formatGoExpression(literal.Type) != "[]ArmourTier" {
		return nil, fmt.Errorf("Dragonfly item.ArmourTiers body changed")
	}
	names := make([]string, 0, len(literal.Elts))
	for _, raw := range literal.Elts {
		value, ok := raw.(*ast.CompositeLit)
		if !ok {
			return nil, fmt.Errorf("Dragonfly item.ArmourTiers contains unsupported value")
		}
		name := formatGoExpression(value.Type)
		if !strings.HasPrefix(name, "ArmourTier") {
			return nil, fmt.Errorf("Dragonfly item.ArmourTiers contains unsupported type %s", name)
		}
		names = append(names, name)
	}
	return names, nil
}

func validateArmourTierType(spec *ast.TypeSpec, typeOf reflect.Type, methods map[string]*ast.FuncDecl) error {
	fields := exportedItemFields(spec)
	wantFields := []string(nil)
	if typeOf.Name() == "ArmourTierLeather" {
		wantFields = []string{"Colour color.RGBA"}
	}
	if !reflect.DeepEqual(fields, wantFields) {
		return fmt.Errorf("Dragonfly item.%s fields changed: %v", typeOf.Name(), fields)
	}
	for name, signature := range map[string]goSignature{
		"BaseDurability":      {Results: "float64"},
		"Toughness":           {Results: "float64"},
		"KnockBackResistance": {Results: "float64"},
		"EnchantmentValue":    {Results: "int"},
		"Name":                {Results: "string"},
	} {
		method := methods[name]
		if method == nil || !valueReceiver(method, typeOf.Name()) || rawParameterTypes(method.Type.Params) != signature.Parameters || rawResultTypes(method.Type.Results) != signature.Results {
			return fmt.Errorf("Dragonfly item.%s.%s signature changed", typeOf.Name(), name)
		}
	}
	return nil
}

func inspectArmourPiece(
	typeOf reflect.Type,
	values []world.Item,
	declaration *ast.TypeSpec,
	methods map[string]*ast.FuncDecl,
	armourTiers []dfitem.ArmourTier,
	toolTiers []dfitem.ToolTier,
	valueTypes []itemValueTypeSpec,
) (armourPieceSpec, itemTypeSpec, error) {
	if fields := exportedItemFields(declaration); !reflect.DeepEqual(fields, []string{"Tier ArmourTier", "Trim ArmourTrim"}) {
		return armourPieceSpec{}, itemTypeSpec{}, fmt.Errorf("Dragonfly item.%s fields changed: %v", typeOf.Name(), fields)
	}
	slotMethod, err := armourSlotMethod(values[0])
	if err != nil {
		return armourPieceSpec{}, itemTypeSpec{}, err
	}
	if err := validateArmourPieceMethods(typeOf.Name(), slotMethod, methods); err != nil {
		return armourPieceSpec{}, itemTypeSpec{}, err
	}
	divisor, err := inspectArmourDurabilityDivisor(methods["DurabilityInfo"])
	if err != nil {
		return armourPieceSpec{}, itemTypeSpec{}, fmt.Errorf("Dragonfly item.%s.DurabilityInfo: %w", typeOf.Name(), err)
	}
	piece := armourPieceSpec{
		Name: typeOf.Name(), SlotMethod: slotMethod, DurabilityDivisor: divisor,
		DefencePoints: make([]float64, len(armourTiers)), RepairItems: make([]string, len(armourTiers)), Smelts: make([]armourSmeltSpec, len(armourTiers)),
	}
	definition := itemTypeSpec{Name: typeOf.Name(), NBT: true, Armour: true}
	byTier := map[reflect.Type]world.Item{}
	for _, value := range values {
		reflected := reflect.ValueOf(value)
		if reflected.Kind() == reflect.Pointer {
			reflected = reflected.Elem()
		}
		tier, ok := reflected.FieldByName("Tier").Interface().(dfitem.ArmourTier)
		if !ok {
			return armourPieceSpec{}, itemTypeSpec{}, fmt.Errorf("Dragonfly item.%s.Tier is not item.ArmourTier", typeOf.Name())
		}
		byTier[reflect.TypeOf(tier)] = value
	}
	for tierIndex, tier := range armourTiers {
		value := byTier[reflect.TypeOf(tier)]
		if value == nil {
			return armourPieceSpec{}, itemTypeSpec{}, fmt.Errorf("Dragonfly item.%s has no %T registry state", typeOf.Name(), tier)
		}
		state, err := inspectItemState(typeOf, value, nil, toolTiers, valueTypes)
		if err != nil {
			return armourPieceSpec{}, itemTypeSpec{}, err
		}
		state.ArmourTier = tierIndex
		definition.States = append(definition.States, state)
		piece.DefencePoints[tierIndex] = value.(dfitem.Armour).DefencePoints()
		repairItem, err := inspectArmourRepairItem(value)
		if err != nil {
			return armourPieceSpec{}, itemTypeSpec{}, err
		}
		piece.RepairItems[tierIndex] = repairItem
		if smeltable, ok := value.(dfitem.Smeltable); ok {
			info := smeltable.SmeltInfo()
			if !info.Product.Empty() {
				product := info.Product.Item()
				productType := reflect.TypeOf(product)
				if productType.Kind() == reflect.Pointer {
					productType = productType.Elem()
				}
				if productType.PkgPath() != "github.com/df-mc/dragonfly/server/item" || productType.Name() == "" {
					return armourPieceSpec{}, itemTypeSpec{}, fmt.Errorf("Dragonfly item.%s tier %T has unsupported smelt product %T", typeOf.Name(), tier, product)
				}
				piece.Smelts[tierIndex] = armourSmeltSpec{
					Product: productType.Name(), Count: info.Product.Count(), Experience: info.Experience, Food: info.Food, Ores: info.Ores,
				}
			}
		}
	}
	return piece, definition, nil
}

func armourSlotMethod(value world.Item) (string, error) {
	switch value.(type) {
	case dfitem.HelmetType:
		return "Helmet", nil
	case dfitem.ChestplateType:
		return "Chestplate", nil
	case dfitem.LeggingsType:
		return "Leggings", nil
	case dfitem.BootsType:
		return "Boots", nil
	default:
		return "", fmt.Errorf("Dragonfly armour item %T has no supported slot interface", value)
	}
}

func validateArmourPieceMethods(name, slotMethod string, methods map[string]*ast.FuncDecl) error {
	wanted := map[string]goSignature{
		"MaxCount":            {Results: "int"},
		"DefencePoints":       {Results: "float64"},
		"Toughness":           {Results: "float64"},
		"KnockBackResistance": {Results: "float64"},
		"EnchantmentValue":    {Results: "int"},
		"DurabilityInfo":      {Results: "DurabilityInfo"},
		"SmeltInfo":           {Results: "SmeltInfo"},
		"RepairableBy":        {Parameters: "Stack", Results: "bool"},
		slotMethod:            {Results: "bool"},
		"WithTrim":            {Parameters: "ArmourTrim", Results: "world.Item"},
		"EncodeItem":          {Results: "string,int16"},
		"DecodeNBT":           {Parameters: "map[string]any", Results: "any"},
		"EncodeNBT":           {Results: "map[string]any"},
	}
	for methodName, signature := range wanted {
		method := methods[methodName]
		if method == nil || !valueReceiver(method, name) || rawParameterTypes(method.Type.Params) != signature.Parameters || rawResultTypes(method.Type.Results) != signature.Results {
			return fmt.Errorf("Dragonfly item.%s.%s signature changed", name, methodName)
		}
	}
	return nil
}

func inspectArmourDurabilityDivisor(method *ast.FuncDecl) (float64, error) {
	if method == nil || method.Body == nil || len(method.Body.List) != 1 {
		return 0, fmt.Errorf("body changed")
	}
	statement, ok := method.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(statement.Results) != 1 {
		return 0, fmt.Errorf("return changed")
	}
	literal, ok := statement.Results[0].(*ast.CompositeLit)
	if !ok || formatGoExpression(literal.Type) != "DurabilityInfo" {
		return 0, fmt.Errorf("return value changed")
	}
	var maximum ast.Expr
	for _, raw := range literal.Elts {
		pair, ok := raw.(*ast.KeyValueExpr)
		if ok && formatGoExpression(pair.Key) == "MaxDurability" {
			maximum = pair.Value
			break
		}
	}
	call, ok := maximum.(*ast.CallExpr)
	if !ok || formatGoExpression(call.Fun) != "int" || len(call.Args) != 1 {
		return 0, fmt.Errorf("maximum durability changed")
	}
	if isArmourBaseDurability(call.Args[0]) {
		return 0, nil
	}
	addition, ok := call.Args[0].(*ast.BinaryExpr)
	if !ok || addition.Op != token.ADD || !isArmourBaseDurability(addition.X) {
		return 0, fmt.Errorf("maximum durability formula changed: %s", formatGoExpression(call.Args[0]))
	}
	division, ok := addition.Y.(*ast.BinaryExpr)
	if !ok || division.Op != token.QUO || !isArmourBaseDurability(division.X) {
		return 0, fmt.Errorf("maximum durability formula changed: %s", formatGoExpression(addition.Y))
	}
	literalDivisor, ok := division.Y.(*ast.BasicLit)
	if !ok || literalDivisor.Kind != token.FLOAT {
		return 0, fmt.Errorf("maximum durability divisor changed")
	}
	divisor, err := strconv.ParseFloat(literalDivisor.Value, 64)
	if err != nil || divisor <= 0 {
		return 0, fmt.Errorf("invalid maximum durability divisor %q", literalDivisor.Value)
	}
	return divisor, nil
}

func isArmourBaseDurability(expression ast.Expr) bool {
	call, ok := expression.(*ast.CallExpr)
	if !ok || len(call.Args) != 0 {
		return false
	}
	method, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || method.Sel.Name != "BaseDurability" {
		return false
	}
	tier, ok := method.X.(*ast.SelectorExpr)
	if !ok || tier.Sel.Name != "Tier" {
		return false
	}
	_, ok = tier.X.(*ast.Ident)
	return ok
}

func inspectArmourRepairItem(value world.Item) (string, error) {
	repairable, ok := value.(dfitem.Repairable)
	if !ok {
		return "", fmt.Errorf("Dragonfly armour item %T is not repairable", value)
	}
	seen := map[string]bool{}
	for _, candidate := range world.Items() {
		typeOf := reflect.TypeOf(candidate)
		if typeOf == nil {
			continue
		}
		if typeOf.Kind() == reflect.Pointer {
			typeOf = typeOf.Elem()
		}
		if typeOf.PkgPath() != "github.com/df-mc/dragonfly/server/item" || typeOf.Name() == "" || seen[typeOf.Name()] {
			continue
		}
		seen[typeOf.Name()] = true
		if repairable.RepairableBy(dfitem.NewStack(candidate, 1)) {
			return typeOf.Name(), nil
		}
	}
	return "", fmt.Errorf("Dragonfly armour item %T has no registered repair item", value)
}

func inspectFireworkItemType(
	typeOf reflect.Type,
	values []world.Item,
	declaration *ast.TypeSpec,
	types map[string]*ast.TypeSpec,
	methods map[string]map[string]*ast.FuncDecl,
	tiers []dfitem.ToolTier,
	valueTypes []itemValueTypeSpec,
) (itemTypeSpec, error) {
	name := typeOf.Name()
	if err := validateFireworkItemTypes(declaration, name, types, methods); err != nil {
		return itemTypeSpec{}, err
	}
	definition := itemTypeSpec{Name: name, NBT: true}
	if name == "Firework" {
		if len(values) != 1 {
			return itemTypeSpec{}, fmt.Errorf("Dragonfly item.Firework registry states changed: %d", len(values))
		}
		state, err := inspectItemState(typeOf, values[0], nil, tiers, valueTypes)
		if err != nil {
			return itemTypeSpec{}, err
		}
		definition.States = []itemStateSpec{state}
		return definition, nil
	}

	colours := findItemValueType(valueTypes, "Colour")
	if colours == nil || len(values) != len(colours.Values) {
		return itemTypeSpec{}, fmt.Errorf("Dragonfly item.FireworkStar registry states changed: %d", len(values))
	}
	definition.Fields = []itemFieldSpec{{Name: "Colour", Kind: itemFieldValue, ValueType: "Colour"}}
	for index, raw := range colours.Values {
		colour, ok := raw.(dfitem.Colour)
		if !ok {
			return itemTypeSpec{}, fmt.Errorf("Dragonfly item.Colour value %d has type %T", index, raw)
		}
		value := dfitem.FireworkStar{FireworkExplosion: dfitem.FireworkExplosion{Colour: colour}}
		state, err := inspectItemState(typeOf, value, nil, tiers, valueTypes)
		if err != nil {
			return itemTypeSpec{}, err
		}
		state.Values = []int{index}
		definition.States = append(definition.States, state)
	}
	return definition, nil
}

func validateFireworkItemTypes(
	spec *ast.TypeSpec,
	name string,
	types map[string]*ast.TypeSpec,
	methods map[string]map[string]*ast.FuncDecl,
) error {
	wantFields := map[string][]string{
		"Firework":     {"Duration time.Duration", "Explosions []FireworkExplosion"},
		"FireworkStar": {"FireworkExplosion"},
	}[name]
	if fields := exportedItemFields(spec); !reflect.DeepEqual(fields, wantFields) {
		return fmt.Errorf("Dragonfly item.%s fields changed: %v", name, fields)
	}
	if fields := exportedItemFields(types["FireworkExplosion"]); !reflect.DeepEqual(fields, []string{
		"Shape FireworkShape", "Colour Colour", "Fade Colour", "Fades bool", "Twinkle bool", "Trail bool",
	}) {
		return fmt.Errorf("Dragonfly item.FireworkExplosion fields changed: %v", fields)
	}
	wantMethods := map[string]map[string]goSignature{
		"Firework": {
			"EncodeNBT":          {Results: "map[string]any"},
			"DecodeNBT":          {Parameters: "map[string]any", Results: "any"},
			"RandomisedDuration": {Results: "time.Duration"},
			"OffHand":            {Results: "bool"},
			"EncodeItem":         {Results: "string,int16"},
		},
		"FireworkStar": {
			"EncodeNBT":  {Results: "map[string]any"},
			"DecodeNBT":  {Parameters: "map[string]any", Results: "any"},
			"EncodeItem": {Results: "string,int16"},
		},
	}[name]
	for methodName, signature := range wantMethods {
		method := methods[name][methodName]
		if method == nil || !valueReceiver(method, name) ||
			rawParameterTypes(method.Type.Params) != signature.Parameters || rawResultTypes(method.Type.Results) != signature.Results {
			return fmt.Errorf("Dragonfly item.%s.%s signature changed", name, methodName)
		}
	}
	for methodName, signature := range map[string]goSignature{
		"EncodeNBT": {Results: "map[string]any"},
		"DecodeNBT": {Parameters: "map[string]any", Results: "any"},
	} {
		method := methods["FireworkExplosion"][methodName]
		if method == nil || !valueReceiver(method, "FireworkExplosion") ||
			rawParameterTypes(method.Type.Params) != signature.Parameters || rawResultTypes(method.Type.Results) != signature.Results {
			return fmt.Errorf("Dragonfly item.FireworkExplosion.%s signature changed", methodName)
		}
	}
	return nil
}

func exportedItemFields(spec *ast.TypeSpec) []string {
	structure, ok := biomeEmptyStruct(spec)
	if !ok {
		return nil
	}
	var fields []string
	for _, field := range structure.Fields.List {
		if len(field.Names) == 0 {
			if name, ok := field.Type.(*ast.Ident); ok && name.IsExported() {
				fields = append(fields, name.Name)
			}
			continue
		}
		for _, name := range field.Names {
			if name.IsExported() {
				fields = append(fields, name.Name+" "+formatGoExpression(field.Type))
			}
		}
	}
	return fields
}

func allItemFields(spec *ast.TypeSpec) []string {
	structure, ok := biomeEmptyStruct(spec)
	if !ok {
		return nil
	}
	var fields []string
	for _, field := range structure.Fields.List {
		for _, name := range field.Names {
			fields = append(fields, name.Name+" "+formatGoExpression(field.Type))
		}
	}
	return fields
}

func validateBookItemType(spec *ast.TypeSpec, name string, methods map[string]*ast.FuncDecl) error {
	structure, ok := biomeEmptyStruct(spec)
	if !ok {
		return fmt.Errorf("Dragonfly item.%s is not a struct", name)
	}
	var fields []string
	for _, field := range structure.Fields.List {
		for _, fieldName := range field.Names {
			if fieldName.IsExported() {
				fields = append(fields, fieldName.Name+" "+formatGoExpression(field.Type))
			}
		}
	}
	want := map[string][]string{
		"BookAndQuill": {"Pages []string"},
		"WrittenBook":  {"Title string", "Author string", "Generation WrittenBookGeneration", "Pages []string"},
	}[name]
	if !reflect.DeepEqual(fields, want) {
		return fmt.Errorf("Dragonfly item.%s fields changed: %v", name, fields)
	}
	wantMethods := map[string]map[string]goSignature{
		"BookAndQuill": {
			"MaxCount": {Results: "int"}, "TotalPages": {Results: "int"},
			"Page":       {Parameters: "int", Results: "string,bool"},
			"DeletePage": {Parameters: "int", Results: "BookAndQuill"},
			"InsertPage": {Parameters: "int,string", Results: "BookAndQuill"},
			"SetPage":    {Parameters: "int,string", Results: "BookAndQuill"},
			"SwapPages":  {Parameters: "int,int", Results: "BookAndQuill"},
			"DecodeNBT":  {Parameters: "map[string]any", Results: "any"},
			"EncodeNBT":  {Results: "map[string]any"},
			"EncodeItem": {Results: "string,int16"},
		},
		"WrittenBook": {
			"MaxCount": {Results: "int"}, "TotalPages": {Results: "int"},
			"Page":       {Parameters: "int", Results: "string,bool"},
			"DecodeNBT":  {Parameters: "map[string]any", Results: "any"},
			"EncodeNBT":  {Results: "map[string]any"},
			"EncodeItem": {Results: "string,int16"},
		},
	}[name]
	for methodName, signature := range wantMethods {
		method := methods[methodName]
		if method == nil || !valueReceiver(method, name) ||
			rawParameterTypes(method.Type.Params) != signature.Parameters || rawResultTypes(method.Type.Results) != signature.Results {
			if method == nil {
				return fmt.Errorf("Dragonfly item.%s.%s missing", name, methodName)
			}
			return fmt.Errorf("Dragonfly item.%s.%s signature changed: (%s) (%s)", name, methodName,
				rawParameterTypes(method.Type.Params), rawResultTypes(method.Type.Results))
		}
	}
	return nil
}

func inspectItemValueTypes(directory string, itemFunctions map[string]*ast.FuncDecl, effects effectSpec) ([]itemValueTypeSpec, error) {
	potionFunctions, err := packageFunctions(filepath.Join(directory, "potion"), "potion")
	if err != nil {
		return nil, err
	}
	soundFunctions, err := packageFunctions(filepath.Join(filepath.Dir(directory), "world", "sound"), "sound")
	if err != nil {
		return nil, err
	}
	specs := []itemValueTypeSpec{
		{GoType: "Colour", CSharpType: "Colour", Container: "Item", Name: "Colour", Values: anySlice(dfitem.Colours())},
		{GoType: "FireworkShape", CSharpType: "FireworkShape", Container: "Item", Name: "FireworkShape", Values: anySlice(dfitem.FireworkShapes())},
		{GoType: "SmithingTemplateType", CSharpType: "SmithingTemplateType", Container: "Item", Name: "SmithingTemplateType", Values: anySlice(dfitem.SmithingTemplates())},
		{GoType: "BannerPatternType", CSharpType: "BannerPatternType", Container: "Item", Name: "BannerPatternType", Values: anySlice(dfitem.BannerPatterns())},
		{GoType: "StewType", CSharpType: "StewType", Container: "Item", Name: "StewType", Values: anySlice(dfitem.StewTypes())},
		{GoType: "SherdType", CSharpType: "SherdType", Container: "Item", Name: "SherdType", Values: anySlice(dfitem.SherdTypes())},
		{GoType: "WrittenBookGeneration", CSharpType: "WrittenBookGeneration", Container: "Item", Name: "WrittenBookGeneration", Values: []any{dfitem.OriginalGeneration(), dfitem.CopyGeneration(), dfitem.CopyOfCopyGeneration()}},
		{GoType: "potion.Potion", CSharpType: "global::Dragonfly.Potion.Value", Container: "Potion", Name: "Value", Values: anySlice(dfpotion.All())},
		{GoType: "sound.Horn", CSharpType: "Sound.Horn", Container: "Sound", Name: "Horn", Values: anySlice(dfsound.GoatHorns())},
		{GoType: "sound.DiscType", CSharpType: "Sound.DiscType", Container: "Sound", Name: "DiscType", Values: anySlice(dfsound.MusicDiscs())},
	}
	collections := map[string]struct {
		functions map[string]*ast.FuncDecl
		name      string
	}{
		"Colour":               {itemFunctions, "Colours"},
		"FireworkShape":        {itemFunctions, "FireworkShapes"},
		"SmithingTemplateType": {itemFunctions, "SmithingTemplates"},
		"BannerPatternType":    {itemFunctions, "BannerPatterns"},
		"StewType":             {itemFunctions, "StewTypes"},
		"SherdType":            {itemFunctions, "SherdTypes"},
		"potion.Potion":        {potionFunctions, "All"},
		"sound.Horn":           {soundFunctions, "GoatHorns"},
		"sound.DiscType":       {soundFunctions, "MusicDiscs"},
	}
	for index := range specs {
		collection, listed := collections[specs[index].GoType]
		var factories []string
		var err error
		if listed {
			factories, err = collectionFactoryNames(collection.functions[collection.name])
		} else {
			factories, err = indexedFactoryNames(itemFunctions, specs[index].GoType, len(specs[index].Values))
			collection.name = specs[index].GoType
		}
		if err != nil {
			return nil, fmt.Errorf("Dragonfly %s: %w", collection.name, err)
		}
		if len(factories) != len(specs[index].Values) {
			return nil, fmt.Errorf("Dragonfly %s AST/live lengths differ: %d/%d", collection.name, len(factories), len(specs[index].Values))
		}
		specs[index].Collection = collection.name
		if specs[index].GoType == "potion.Potion" {
			from := potionFunctions["From"]
			if from == nil || goFunctionSignature(from) != (goSignature{Parameters: "int32", Results: "Potion"}) {
				return nil, fmt.Errorf("Dragonfly potion.From signature changed")
			}
			specs[index].From = true
		}
		specs[index].Factories = factories
		methods, err := inspectItemValueMethods(specs[index], effects)
		if err != nil {
			return nil, err
		}
		specs[index].Methods = methods
	}
	return specs, nil
}

func indexedFactoryNames(functions map[string]*ast.FuncDecl, resultType string, count int) ([]string, error) {
	names := make([]string, count)
	for _, function := range functions {
		if function.Recv != nil || function.Type.Params == nil || function.Type.Params.NumFields() != 0 ||
			function.Type.Results == nil || len(function.Type.Results.List) != 1 ||
			formatGoExpression(function.Type.Results.List[0].Type) != resultType || function.Body == nil || len(function.Body.List) != 1 {
			continue
		}
		statement, ok := function.Body.List[0].(*ast.ReturnStmt)
		if !ok || len(statement.Results) != 1 {
			continue
		}
		literal, ok := statement.Results[0].(*ast.CompositeLit)
		if !ok || formatGoExpression(literal.Type) != resultType || len(literal.Elts) != 1 {
			continue
		}
		indexLiteral, ok := literal.Elts[0].(*ast.BasicLit)
		if !ok || indexLiteral.Kind != token.INT {
			continue
		}
		value, err := strconv.Atoi(indexLiteral.Value)
		if err != nil || value < 0 || value >= count || names[value] != "" {
			return nil, fmt.Errorf("Dragonfly %s factory index changed", resultType)
		}
		names[value] = function.Name.Name
	}
	for index, name := range names {
		if name == "" {
			return nil, fmt.Errorf("Dragonfly %s factory %d missing", resultType, index)
		}
	}
	return names, nil
}

func inspectItemValueMethods(spec itemValueTypeSpec, effects effectSpec) ([]itemValueMethodSpec, error) {
	methods := map[string][]struct {
		Name       string
		ReturnType string
	}{
		"Colour": {
			{Name: "RGBA", ReturnType: "Color.RGBA"},
			{Name: "SignRGBA", ReturnType: "Color.RGBA"},
			{Name: "String", ReturnType: "string"},
			{Name: "SilverString", ReturnType: "string"},
			{Name: "Uint8", ReturnType: "byte"},
		},
		"FireworkShape": {
			{Name: "Uint8", ReturnType: "byte"},
			{Name: "Name", ReturnType: "string"},
			{Name: "String", ReturnType: "string"},
		},
		"SmithingTemplateType": {{Name: "String", ReturnType: "string"}},
		"BannerPatternType": {
			{Name: "Uint8", ReturnType: "byte"},
			{Name: "String", ReturnType: "string"},
		},
		"StewType": {
			{Name: "Uint8", ReturnType: "byte"},
			{Name: "Effects", ReturnType: "IReadOnlyList<Effect.Value>"},
		},
		"SherdType": {{Name: "String", ReturnType: "string"}, {Name: "Uint8", ReturnType: "byte"}},
		"WrittenBookGeneration": {
			{Name: "Uint8", ReturnType: "byte"},
			{Name: "String", ReturnType: "string"},
		},
		"potion.Potion": {
			{Name: "Uint8", ReturnType: "byte"},
			{Name: "Effects", ReturnType: "IReadOnlyList<Effect.Value>"},
		},
		"sound.Horn": {{Name: "Uint8", ReturnType: "byte"}, {Name: "Name", ReturnType: "string"}},
		"sound.DiscType": {
			{Name: "Uint8", ReturnType: "byte"},
			{Name: "String", ReturnType: "string"},
			{Name: "DisplayName", ReturnType: "string"},
			{Name: "Author", ReturnType: "string"},
		},
	}[spec.GoType]
	result := make([]itemValueMethodSpec, 0, len(methods))
	for _, method := range methods {
		generated := itemValueMethodSpec{Name: method.Name, ReturnType: method.ReturnType}
		if spec.GoType == "potion.Potion" {
			switch method.Name {
			case "Uint8":
				generated.Default = "unchecked((byte)_value)"
			case "Effects":
				generated.Default = "Array.Empty<Effect.Value>()"
			}
		}
		for _, value := range spec.Values {
			call := reflect.ValueOf(value).MethodByName(method.Name)
			if !call.IsValid() || call.Type().NumIn() != 0 || call.Type().NumOut() != 1 {
				return nil, fmt.Errorf("Dragonfly %s.%s signature changed", spec.GoType, method.Name)
			}
			outputs := call.Call(nil)
			formatted, err := csharpItemValueResult(outputs[0].Interface(), method.ReturnType, effects)
			if err != nil {
				return nil, fmt.Errorf("Dragonfly %s.%s: %w", spec.GoType, method.Name, err)
			}
			generated.Results = append(generated.Results, formatted)
		}
		result = append(result, generated)
	}
	return result, nil
}

func csharpItemValueResult(value any, returnType string, effects effectSpec) (string, error) {
	switch returnType {
	case "byte":
		value, ok := value.(uint8)
		if !ok {
			return "", fmt.Errorf("result is %T, not uint8", value)
		}
		return strconv.Itoa(int(value)), nil
	case "string":
		value, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("result is %T, not string", value)
		}
		return strconv.Quote(value), nil
	case "Color.RGBA":
		value, ok := value.(color.RGBA)
		if !ok {
			return "", fmt.Errorf("result is %T, not color.RGBA", value)
		}
		return fmt.Sprintf("new Color.RGBA(%d, %d, %d, %d)", value.R, value.G, value.B, value.A), nil
	case "IReadOnlyList<Effect.Value>":
		return csharpEffectResult(value, effects)
	default:
		return "", fmt.Errorf("unsupported C# return type %s", returnType)
	}
}

func packageFunctions(directory, packageName string) (map[string]*ast.FuncDecl, error) {
	packages, err := parser.ParseDir(token.NewFileSet(), directory, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, err
	}
	pkg, ok := packages[packageName]
	if !ok {
		return nil, fmt.Errorf("Dragonfly %s package not found", packageName)
	}
	functions := map[string]*ast.FuncDecl{}
	for _, file := range pkg.Files {
		for _, declaration := range file.Decls {
			if function, ok := declaration.(*ast.FuncDecl); ok && function.Recv == nil {
				functions[function.Name.Name] = function
			}
		}
	}
	return functions, nil
}

func collectionFactoryNames(function *ast.FuncDecl) ([]string, error) {
	if function == nil || function.Body == nil || len(function.Body.List) != 1 {
		return nil, fmt.Errorf("collection function body changed")
	}
	statement, ok := function.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(statement.Results) != 1 {
		return nil, fmt.Errorf("collection return changed")
	}
	literal, ok := statement.Results[0].(*ast.CompositeLit)
	if !ok {
		return nil, fmt.Errorf("collection literal changed")
	}
	names := make([]string, 0, len(literal.Elts))
	for _, raw := range literal.Elts {
		call, ok := raw.(*ast.CallExpr)
		if !ok || len(call.Args) != 0 {
			return nil, fmt.Errorf("collection value changed")
		}
		name, ok := call.Fun.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("collection factory changed")
		}
		names = append(names, name.Name)
	}
	return names, nil
}

func anySlice[T any](values []T) []any {
	result := make([]any, len(values))
	for index := range values {
		result[index] = values[index]
	}
	return result
}

func findItemValueType(types []itemValueTypeSpec, goType string) *itemValueTypeSpec {
	for index := range types {
		if types[index].GoType == goType {
			return &types[index]
		}
	}
	return nil
}

func inspectToolTierFields(spec *ast.TypeSpec) ([]parameter, error) {
	structure, ok := biomeEmptyStruct(spec)
	if !ok {
		return nil, fmt.Errorf("Dragonfly item.ToolTier is not a struct")
	}
	var fields []parameter
	for _, field := range structure.Fields.List {
		identifier, ok := field.Type.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("Dragonfly item.ToolTier has unsupported field type %s", formatGoExpression(field.Type))
		}
		typeName, ok := map[string]string{"int": "int", "float64": "double", "string": "string"}[identifier.Name]
		if !ok {
			return nil, fmt.Errorf("Dragonfly item.ToolTier has unsupported field type %s", identifier.Name)
		}
		for _, name := range field.Names {
			if name.IsExported() {
				fields = append(fields, parameter{Name: name.Name, Type: typeName})
			}
		}
	}
	want := []parameter{
		{Name: "HarvestLevel", Type: "int"},
		{Name: "BaseMiningEfficiency", Type: "double"},
		{Name: "BaseAttackDamage", Type: "double"},
		{Name: "EnchantmentValue", Type: "int"},
		{Name: "Durability", Type: "int"},
		{Name: "Name", Type: "string"},
	}
	if !reflect.DeepEqual(fields, want) {
		return nil, fmt.Errorf("Dragonfly item.ToolTier fields changed: %v", fields)
	}
	return fields, nil
}

func inspectToolTierVariables(function *ast.FuncDecl, variables map[string]bool) ([]string, error) {
	if function == nil || function.Type.Params == nil || function.Type.Params.NumFields() != 0 ||
		function.Type.Results == nil || len(function.Type.Results.List) != 1 ||
		formatGoExpression(function.Type.Results.List[0].Type) != "[]ToolTier" || function.Body == nil || len(function.Body.List) != 1 {
		return nil, fmt.Errorf("Dragonfly item.ToolTiers signature/body changed")
	}
	statement, ok := function.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(statement.Results) != 1 {
		return nil, fmt.Errorf("Dragonfly item.ToolTiers body changed")
	}
	literal, ok := statement.Results[0].(*ast.CompositeLit)
	if !ok || formatGoExpression(literal.Type) != "[]ToolTier" {
		return nil, fmt.Errorf("Dragonfly item.ToolTiers body changed")
	}
	names := make([]string, 0, len(literal.Elts))
	for _, raw := range literal.Elts {
		name, ok := raw.(*ast.Ident)
		if !ok || !strings.HasPrefix(name.Name, "ToolTier") || !variables[name.Name] {
			return nil, fmt.Errorf("Dragonfly item.ToolTiers contains unsupported value")
		}
		names = append(names, name.Name)
	}
	return names, nil
}

func inspectItemFields(spec *ast.TypeSpec, name string, valueTypes []itemValueTypeSpec) ([]itemFieldSpec, bool, error) {
	structure, ok := biomeEmptyStruct(spec)
	if !ok {
		return nil, false, nil
	}
	var fields []itemFieldSpec
	var kind *itemFieldKind
	for _, field := range structure.Fields.List {
		for _, fieldName := range field.Names {
			if !fieldName.IsExported() {
				continue
			}
			goType := formatGoExpression(field.Type)
			fieldKind := itemFieldBool
			valueType := ""
			switch goType {
			case "bool":
			case "ToolTier":
				fieldKind = itemFieldToolTier
			default:
				if findItemValueType(valueTypes, goType) == nil {
					return nil, false, nil
				}
				fieldKind = itemFieldValue
				valueType = goType
			}
			if kind != nil && *kind != fieldKind {
				return nil, false, nil
			}
			value := fieldKind
			kind = &value
			fields = append(fields, itemFieldSpec{Name: fieldName.Name, Kind: fieldKind, ValueType: valueType})
		}
	}
	if len(fields) > 16 || (len(fields) > 1 && (fields[0].Kind == itemFieldToolTier || fields[0].Kind == itemFieldValue)) {
		return nil, false, nil
	}
	return fields, true, nil
}

func inspectItemState(typeOf reflect.Type, value world.Item, fields []itemFieldSpec, tiers []dfitem.ToolTier, valueTypes []itemValueTypeSpec) (itemStateSpec, error) {
	identifier, metadata := value.EncodeItem()
	capability := itemCapabilitySpec{MaxCount: 64, MaxDurability: -1, AttackDamage: 1}
	if counter, ok := value.(dfitem.MaxCounter); ok {
		capability.MaxCount = counter.MaxCount()
	}
	if durable, ok := value.(dfitem.Durable); ok {
		info := durable.DurabilityInfo()
		capability.MaxDurability = info.MaxDurability
		capability.Persistent = info.Persistent
		broken := info.BrokenItem()
		capability.BrokenCount = broken.Count()
		if brokenItem := broken.Item(); brokenItem != nil {
			brokenIdentifier, brokenMetadata := brokenItem.EncodeItem()
			capability.BrokenIdentifier = brokenIdentifier
			capability.BrokenMetadata = int(brokenMetadata)
		}
	}
	if weapon, ok := value.(dfitem.Weapon); ok {
		capability.AttackDamage = weapon.AttackDamage() + 1
	}
	_, repairable := value.(dfitem.Repairable)
	_, enchantedBook := value.(dfitem.EnchantedBook)
	capability.AllowsAnvilCost = repairable || enchantedBook
	if fuel, ok := value.(dfitem.Fuel); ok {
		info := fuel.FuelInfo()
		capability.Fuel = true
		capability.FuelDuration = info.Duration
		capability.FuelCount = info.Residue.Count()
		if residue := info.Residue.Item(); residue != nil {
			residueIdentifier, residueMetadata := residue.EncodeItem()
			capability.FuelIdentifier = residueIdentifier
			capability.FuelMetadata = int(residueMetadata)
		}
	}
	state := itemStateSpec{Identifier: identifier, Metadata: int(metadata), ToolTier: -1, ArmourTier: -1, Capability: capability}
	reflected := reflect.ValueOf(value)
	if reflected.Kind() == reflect.Pointer {
		reflected = reflected.Elem()
	}
	for _, field := range fields {
		reflectedField := reflected.FieldByName(field.Name)
		if !reflectedField.IsValid() {
			return itemStateSpec{}, fmt.Errorf("Dragonfly item.%s field %s missing from live type", typeOf.Name(), field.Name)
		}
		switch field.Kind {
		case itemFieldBool:
			state.Bools = append(state.Bools, reflectedField.Bool())
		case itemFieldToolTier:
			tier, ok := reflectedField.Interface().(dfitem.ToolTier)
			if !ok {
				return itemStateSpec{}, fmt.Errorf("Dragonfly item.%s.%s is not item.ToolTier", typeOf.Name(), field.Name)
			}
			for index, candidate := range tiers {
				if tier == candidate {
					state.ToolTier = index
					break
				}
			}
			if state.ToolTier < 0 {
				return itemStateSpec{}, fmt.Errorf("Dragonfly item.%s has unregistered ToolTier %+v", typeOf.Name(), tier)
			}
		case itemFieldValue:
			valueType := findItemValueType(valueTypes, field.ValueType)
			if valueType == nil {
				return itemStateSpec{}, fmt.Errorf("Dragonfly item.%s.%s has unknown value type %s", typeOf.Name(), field.Name, field.ValueType)
			}
			valueIndex := -1
			for index, candidate := range valueType.Values {
				if reflect.DeepEqual(reflectedField.Interface(), candidate) {
					valueIndex = index
					break
				}
			}
			if valueIndex < 0 {
				return itemStateSpec{}, fmt.Errorf("Dragonfly item.%s.%s has unregistered value %+v", typeOf.Name(), field.Name, reflectedField.Interface())
			}
			state.Values = append(state.Values, valueIndex)
		}
	}
	return state, nil
}

func completeItemType(typeOf reflect.Type, definition itemTypeSpec, tiers []dfitem.ToolTier, valueTypes []itemValueTypeSpec) bool {
	seen := map[string]bool{}
	for _, state := range definition.States {
		key := strconv.Itoa(state.ToolTier) + ":"
		for _, value := range state.Values {
			key += strconv.Itoa(value) + ","
		}
		for _, value := range state.Bools {
			key += strconv.FormatBool(value) + ","
		}
		if seen[key] {
			return false
		}
		seen[key] = true
	}
	if len(definition.Fields) == 0 {
		if len(definition.States) != 1 {
			return false
		}
		zero, ok := reflect.Zero(typeOf).Interface().(world.Item)
		if !ok {
			return false
		}
		identifier, metadata := zero.EncodeItem()
		return identifier == definition.States[0].Identifier && int(metadata) == definition.States[0].Metadata
	}
	if definition.Fields[0].Kind == itemFieldToolTier {
		if len(definition.States) != len(tiers) {
			return false
		}
		for index := range tiers {
			if !seen[strconv.Itoa(index)+":"] {
				return false
			}
		}
		return true
	}
	if definition.Fields[0].Kind == itemFieldValue {
		valueType := findItemValueType(valueTypes, definition.Fields[0].ValueType)
		if valueType == nil || len(definition.States) != len(valueType.Values) {
			return false
		}
		for index := range valueType.Values {
			if !seen["-1:"+strconv.Itoa(index)+","] {
				return false
			}
		}
		return true
	}
	return len(definition.States) == 1<<len(definition.Fields)
}

func inspectGameModes(path string) (gameModeSpec, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return gameModeSpec{}, err
	}
	types := map[string]*ast.TypeSpec{}
	variables := map[string]string{}
	functions := map[string]*ast.FuncDecl{}
	var registry *ast.CompositeLit
	for _, declaration := range file.Decls {
		switch value := declaration.(type) {
		case *ast.GenDecl:
			for _, raw := range value.Specs {
				switch spec := raw.(type) {
				case *ast.TypeSpec:
					types[spec.Name.Name] = spec
				case *ast.ValueSpec:
					if identifier, ok := spec.Type.(*ast.Ident); ok {
						for _, name := range spec.Names {
							if name.IsExported() {
								variables[name.Name] = identifier.Name
							}
						}
					}
					for index, name := range spec.Names {
						if name.Name != "gameModeReg" || index >= len(spec.Values) {
							continue
						}
						call, ok := spec.Values[index].(*ast.CallExpr)
						if ok && len(call.Args) == 1 {
							registry, _ = call.Args[0].(*ast.CompositeLit)
						}
					}
				}
			}
		case *ast.FuncDecl:
			if value.Recv == nil && value.Name.IsExported() {
				functions[value.Name.Name] = value
			}
		}
	}
	methods, err := inspectGameModeInterface(types["GameMode"])
	if err != nil {
		return gameModeSpec{}, err
	}
	if !validGameModeLookupFunction(functions["GameModeByID"], true) {
		return gameModeSpec{}, fmt.Errorf("Dragonfly world.GameModeByID signature changed")
	}
	if !validGameModeLookupFunction(functions["GameModeID"], false) {
		return gameModeSpec{}, fmt.Errorf("Dragonfly world.GameModeID signature changed")
	}
	entries, err := inspectGameModeRegistry(registry)
	if err != nil {
		return gameModeSpec{}, err
	}
	if len(entries) != len(gameModeVariableNames) {
		return gameModeSpec{}, fmt.Errorf("Dragonfly game mode registry has %d entries, want exactly %d", len(entries), len(gameModeVariableNames))
	}
	live := map[string]world.GameMode{
		"GameModeSurvival":  world.GameModeSurvival,
		"GameModeCreative":  world.GameModeCreative,
		"GameModeAdventure": world.GameModeAdventure,
		"GameModeSpectator": world.GameModeSpectator,
	}
	for _, name := range gameModeVariableNames {
		if variables[name] == "" {
			return gameModeSpec{}, fmt.Errorf("Dragonfly world.%s variable declaration not found", name)
		}
		structure, ok := biomeEmptyStruct(types[variables[name]])
		if !ok || len(structure.Fields.List) != 0 {
			return gameModeSpec{}, fmt.Errorf("Dragonfly world.%s concrete type is not an empty private struct", name)
		}
	}

	result := gameModeSpec{Methods: methods, Modes: make([]gameModeValue, 0, len(entries))}
	for _, entry := range entries {
		mode := live[entry.Name]
		if mode == nil {
			return gameModeSpec{}, fmt.Errorf("Dragonfly game mode registry contains unexpected %s", entry.Name)
		}
		lookedUp, ok := world.GameModeByID(entry.ID)
		if !ok || lookedUp != mode {
			return gameModeSpec{}, fmt.Errorf("Dragonfly live game mode ID %d does not resolve to %s", entry.ID, entry.Name)
		}
		id, ok := world.GameModeID(mode)
		if !ok || id != entry.ID {
			return gameModeSpec{}, fmt.Errorf("Dragonfly live %s reverse ID is %d, %v", entry.Name, id, ok)
		}
		typeOf := reflect.TypeOf(mode)
		if typeOf.Name() != variables[entry.Name] {
			return gameModeSpec{}, fmt.Errorf("Dragonfly live %s type is %s, want %s", entry.Name, typeOf.Name(), variables[entry.Name])
		}
		capabilities, err := liveGameModeCapabilities(mode, methods)
		if err != nil {
			return gameModeSpec{}, fmt.Errorf("Dragonfly live %s: %w", entry.Name, err)
		}
		result.Modes = append(result.Modes, gameModeValue{
			Name: entry.Name, PrivateType: variables[entry.Name], ID: entry.ID, Capabilities: capabilities,
		})
	}
	unknown := math.MaxInt32
	for _, entry := range entries {
		if entry.ID == unknown {
			unknown--
		}
	}
	fallback, ok := world.GameModeByID(unknown)
	if ok || fallback != world.GameModeSurvival {
		return gameModeSpec{}, fmt.Errorf("Dragonfly unknown game mode fallback changed")
	}
	return result, nil
}

func inspectGameModeInterface(spec *ast.TypeSpec) ([]string, error) {
	if spec == nil {
		return nil, fmt.Errorf("Dragonfly world.GameMode interface not found")
	}
	interfaceType, ok := spec.Type.(*ast.InterfaceType)
	if !ok {
		return nil, fmt.Errorf("Dragonfly world.GameMode is not an interface")
	}
	var methods []string
	for _, field := range interfaceType.Methods.List {
		function, ok := field.Type.(*ast.FuncType)
		if !ok || len(field.Names) != 1 || function.Params == nil || function.Params.NumFields() != 0 || function.Results == nil || len(function.Results.List) != 1 {
			return nil, fmt.Errorf("Dragonfly world.GameMode contains a non-boolean method")
		}
		result, ok := function.Results.List[0].Type.(*ast.Ident)
		if !ok || result.Name != "bool" {
			return nil, fmt.Errorf("Dragonfly world.GameMode.%s does not return bool", field.Names[0].Name)
		}
		methods = append(methods, field.Names[0].Name)
	}
	if !reflect.DeepEqual(methods, gameModeMethodNames) {
		return nil, fmt.Errorf("Dragonfly world.GameMode methods changed: got %v", methods)
	}
	return methods, nil
}

func validGameModeLookupFunction(function *ast.FuncDecl, byID bool) bool {
	if function == nil || function.Type.Params == nil || function.Type.Results == nil || function.Type.Params.NumFields() != 1 || function.Type.Results.NumFields() != 2 {
		return false
	}
	parameterType, resultType := "GameMode", "int"
	if byID {
		parameterType, resultType = "int", "GameMode"
	}
	return formatGoExpression(function.Type.Params.List[0].Type) == parameterType &&
		formatGoExpression(function.Type.Results.List[0].Type) == resultType &&
		formatGoExpression(function.Type.Results.List[1].Type) == "bool"
}

func inspectGameModeRegistry(registry *ast.CompositeLit) ([]gameModeValue, error) {
	if registry == nil {
		return nil, fmt.Errorf("Dragonfly gameModeReg map literal not found")
	}
	mapType, ok := registry.Type.(*ast.MapType)
	if !ok || formatGoExpression(mapType.Key) != "int" || formatGoExpression(mapType.Value) != "GameMode" {
		return nil, fmt.Errorf("Dragonfly gameModeReg is not map[int]GameMode")
	}
	ids := map[int]bool{}
	names := map[string]bool{}
	entries := make([]gameModeValue, 0, len(registry.Elts))
	for _, raw := range registry.Elts {
		entry, ok := raw.(*ast.KeyValueExpr)
		if !ok {
			return nil, fmt.Errorf("Dragonfly gameModeReg contains an unsupported entry")
		}
		key, keyOK := entry.Key.(*ast.BasicLit)
		name, nameOK := entry.Value.(*ast.Ident)
		if !keyOK || key.Kind != token.INT || !nameOK || !name.IsExported() {
			return nil, fmt.Errorf("Dragonfly gameModeReg contains an unsupported entry")
		}
		id, err := strconv.ParseInt(key.Value, 0, 32)
		if err != nil || id < 0 || id > math.MaxInt32 || ids[int(id)] || names[name.Name] {
			return nil, fmt.Errorf("Dragonfly gameModeReg contains invalid ID/name %s:%s", key.Value, name.Name)
		}
		ids[int(id)], names[name.Name] = true, true
		entries = append(entries, gameModeValue{Name: name.Name, ID: int(id)})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	if len(entries) != len(gameModeVariableNames) {
		return nil, fmt.Errorf("Dragonfly game mode registry has %d entries, want exactly %d", len(entries), len(gameModeVariableNames))
	}
	for index, name := range gameModeVariableNames {
		if entries[index].ID != index || entries[index].Name != name {
			return nil, fmt.Errorf("Dragonfly game mode registry changed at ID %d: got %s", index, entries[index].Name)
		}
	}
	return entries, nil
}

func liveGameModeCapabilities(mode world.GameMode, methods []string) ([]bool, error) {
	value := reflect.ValueOf(mode)
	capabilities := make([]bool, len(methods))
	for index, name := range methods {
		method := value.MethodByName(name)
		if !method.IsValid() {
			return nil, fmt.Errorf("method %s not found", name)
		}
		results := method.Call(nil)
		if len(results) != 1 || results[0].Kind() != reflect.Bool {
			return nil, fmt.Errorf("method %s does not return bool", name)
		}
		capabilities[index] = results[0].Bool()
	}
	return capabilities, nil
}

func inspectPlayerGameModeMethods(path string) ([]commandMethod, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	found := map[string]commandMethod{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !pointerReceiver(function, "Player") || (function.Name.Name != "SetGameMode" && function.Name.Name != "GameMode") {
			continue
		}
		method := commandMethod{Name: function.Name.Name}
		switch function.Name.Name {
		case "SetGameMode":
			if function.Type.Params == nil || len(function.Type.Params.List) != 1 || len(function.Type.Params.List[0].Names) != 1 || function.Type.Results != nil {
				return nil, fmt.Errorf("player.Player.SetGameMode signature changed")
			}
			field := function.Type.Params.List[0]
			if field.Names[0].Name != "mode" || formatGoExpression(field.Type) != "world.GameMode" {
				return nil, fmt.Errorf("player.Player.SetGameMode signature changed")
			}
			method.ReturnType = "void"
			method.Parameters = []parameter{{Name: "mode", Type: "World.GameMode"}}
		case "GameMode":
			if function.Type.Params == nil || function.Type.Params.NumFields() != 0 || function.Type.Results == nil || len(function.Type.Results.List) != 1 || formatGoExpression(function.Type.Results.List[0].Type) != "world.GameMode" {
				return nil, fmt.Errorf("player.Player.GameMode signature changed")
			}
			method.ReturnType = "World.GameMode"
		}
		found[method.Name] = method
	}
	result := make([]commandMethod, 0, 2)
	for _, name := range []string{"SetGameMode", "GameMode"} {
		method, ok := found[name]
		if !ok {
			return nil, fmt.Errorf("Dragonfly player.Player has no %s method", name)
		}
		result = append(result, method)
	}
	return result, nil
}

func inspectParticles(directory, instrumentPath, colourPath string) (particleSpec, error) {
	packages, err := parser.ParseDir(token.NewFileSet(), directory, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		return particleSpec{}, err
	}
	pkg, ok := packages["particle"]
	if !ok {
		return particleSpec{}, fmt.Errorf("Dragonfly particle package not found")
	}
	declarations := map[string]*ast.TypeSpec{}
	for _, file := range pkg.Files {
		for _, declaration := range file.Decls {
			gen, ok := declaration.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, raw := range gen.Specs {
				spec, ok := raw.(*ast.TypeSpec)
				if ok {
					declarations[spec.Name.Name] = spec
				}
			}
		}
	}
	marker, ok := declarations["particle"]
	markerStruct, markerIsStruct := biomeEmptyStruct(marker)
	if !ok || !markerIsStruct || len(markerStruct.Fields.List) != 0 {
		return particleSpec{}, fmt.Errorf("Dragonfly particle marker is not an empty private struct")
	}

	exported := map[string]*ast.TypeSpec{}
	for name, declaration := range declarations {
		if ast.IsExported(name) {
			exported[name] = declaration
		}
	}
	if len(exported) != len(particleKindNames) {
		return particleSpec{}, fmt.Errorf("Dragonfly particle package has %d exported types, want exactly %d", len(exported), len(particleKindNames))
	}

	result := particleSpec{Types: make([]particleType, 0, len(particleKindNames))}
	for kind, name := range particleKindNames {
		declaration := exported[name]
		if declaration == nil {
			return particleSpec{}, fmt.Errorf("Dragonfly particle.%s declaration not found", name)
		}
		structure, ok := declaration.Type.(*ast.StructType)
		if !ok {
			return particleSpec{}, fmt.Errorf("Dragonfly particle.%s is not a concrete struct", name)
		}
		fields, err := inspectParticleFields(name, structure)
		if err != nil {
			return particleSpec{}, err
		}
		result.Types = append(result.Types, particleType{Name: name, Kind: uint32(kind), Fields: fields})
	}

	result.Instruments, err = inspectInstruments(instrumentPath)
	if err != nil {
		return particleSpec{}, err
	}
	result.RGBAFields, err = inspectRGBA(colourPath)
	if err != nil {
		return particleSpec{}, err
	}
	return result, nil
}

func inspectParticleFields(name string, structure *ast.StructType) ([]parameter, error) {
	marker := false
	var fields []parameter
	usedSlots := map[string]string{}
	for _, field := range structure.Fields.List {
		if len(field.Names) == 0 {
			identifier, ok := field.Type.(*ast.Ident)
			if !ok || identifier.Name != "particle" || marker {
				return nil, fmt.Errorf("Dragonfly particle.%s does not embed exactly one private particle marker", name)
			}
			marker = true
			continue
		}
		typeName, slot, ok := particleCSharpType(field.Type)
		if !ok {
			return nil, fmt.Errorf("Dragonfly particle.%s has unsupported field type %s", name, formatGoExpression(field.Type))
		}
		for _, fieldName := range field.Names {
			if !fieldName.IsExported() {
				return nil, fmt.Errorf("Dragonfly particle.%s field %s is not exported", name, fieldName.Name)
			}
			if previous, exists := usedSlots[slot]; exists {
				return nil, fmt.Errorf("Dragonfly particle.%s fields %s and %s share encoded slot %s", name, previous, fieldName.Name, slot)
			}
			usedSlots[slot] = fieldName.Name
			fields = append(fields, parameter{Name: fieldName.Name, Type: typeName})
		}
	}
	if !marker {
		return nil, fmt.Errorf("Dragonfly particle.%s does not embed the private particle marker", name)
	}
	return fields, nil
}

func particleCSharpType(expression ast.Expr) (typeName, slot string, ok bool) {
	switch value := expression.(type) {
	case *ast.Ident:
		mapped, exists := map[string]struct{ typeName, slot string }{
			"bool": {"bool", "data"},
			"int":  {"int", "pitch"},
		}[value.Name]
		return mapped.typeName, mapped.slot, exists
	case *ast.SelectorExpr:
		packageName, ok := value.X.(*ast.Ident)
		if !ok {
			return "", "", false
		}
		mapped, exists := map[string]struct{ typeName, slot string }{
			"color.RGBA":       {"Color.RGBA", "colour"},
			"cube.Face":        {"Cube.Face", "data"},
			"cube.Pos":         {"Cube.Pos", "diff"},
			"sound.Instrument": {"Sound.Instrument", "data"},
			"world.Block":      {"World.Block", "block"},
		}[packageName.Name+"."+value.Sel.Name]
		return mapped.typeName, mapped.slot, exists
	default:
		return "", "", false
	}
}

func inspectInstruments(path string) ([]instrumentSpec, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	types := map[string]*ast.TypeSpec{}
	functions := map[string]*ast.FuncDecl{}
	for _, declaration := range file.Decls {
		switch value := declaration.(type) {
		case *ast.GenDecl:
			for _, raw := range value.Specs {
				if spec, ok := raw.(*ast.TypeSpec); ok {
					types[spec.Name.Name] = spec
				}
			}
		case *ast.FuncDecl:
			if value.Recv == nil && value.Name.IsExported() {
				functions[value.Name.Name] = value
			}
		}
	}
	if !validInstrumentTypes(types["Instrument"], types["instrument"]) {
		return nil, fmt.Errorf("Dragonfly sound.Instrument is no longer an opaque int32-backed value")
	}
	if len(functions) != len(instrumentNames) {
		return nil, fmt.Errorf("Dragonfly instrument package has %d exported functions, want exactly %d", len(functions), len(instrumentNames))
	}
	result := make([]instrumentSpec, 0, len(instrumentNames))
	for id, name := range instrumentNames {
		function := functions[name]
		value, ok := instrumentFunctionID(function)
		if !ok || value != uint64(id) {
			return nil, fmt.Errorf("Dragonfly sound.%s is not Instrument{%d}", name, id)
		}
		result = append(result, instrumentSpec{Name: name, ID: uint32(id)})
	}
	return result, nil
}

func validInstrumentTypes(exported, private *ast.TypeSpec) bool {
	if exported == nil || private == nil {
		return false
	}
	structure, ok := biomeEmptyStruct(exported)
	if !ok || len(structure.Fields.List) != 1 {
		return false
	}
	field := structure.Fields.List[0]
	marker, ok := field.Type.(*ast.Ident)
	backing, backingOK := private.Type.(*ast.Ident)
	return len(field.Names) == 0 && ok && marker.Name == "instrument" && backingOK && backing.Name == "int32"
}

func instrumentFunctionID(function *ast.FuncDecl) (uint64, bool) {
	if function == nil || function.Type.Params == nil || function.Type.Params.NumFields() != 0 || function.Type.Results == nil || len(function.Type.Results.List) != 1 || function.Body == nil || len(function.Body.List) != 1 {
		return 0, false
	}
	result, ok := function.Type.Results.List[0].Type.(*ast.Ident)
	if !ok || result.Name != "Instrument" {
		return 0, false
	}
	returnStatement, ok := function.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(returnStatement.Results) != 1 {
		return 0, false
	}
	composite, ok := returnStatement.Results[0].(*ast.CompositeLit)
	if !ok || len(composite.Elts) != 1 {
		return 0, false
	}
	typeName, ok := composite.Type.(*ast.Ident)
	literal, literalOK := composite.Elts[0].(*ast.BasicLit)
	if !ok || typeName.Name != "Instrument" || !literalOK || literal.Kind != token.INT {
		return 0, false
	}
	value, err := strconv.ParseUint(literal.Value, 0, 32)
	return value, err == nil
}

func inspectRGBA(path string) ([]parameter, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	var declaration *ast.TypeSpec
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.TypeSpec)
		if ok && spec.Name.Name == "RGBA" {
			declaration = spec
			return false
		}
		return true
	})
	structure, ok := biomeEmptyStruct(declaration)
	if !ok {
		return nil, fmt.Errorf("image/color.RGBA is not a struct")
	}
	var fields []parameter
	for _, field := range structure.Fields.List {
		identifier, ok := field.Type.(*ast.Ident)
		if !ok || identifier.Name != "uint8" {
			return nil, fmt.Errorf("image/color.RGBA contains non-uint8 fields")
		}
		for _, name := range field.Names {
			fields = append(fields, parameter{Name: name.Name, Type: "byte"})
		}
	}
	want := []parameter{{Name: "R", Type: "byte"}, {Name: "G", Type: "byte"}, {Name: "B", Type: "byte"}, {Name: "A", Type: "byte"}}
	if !reflect.DeepEqual(fields, want) {
		return nil, fmt.Errorf("image/color.RGBA fields changed: got %v", fields)
	}
	return fields, nil
}

func validateLiquidFields(spec *ast.TypeSpec, name string) error {
	if spec == nil {
		return fmt.Errorf("Dragonfly block.%s declaration not found", name)
	}
	structure, ok := spec.Type.(*ast.StructType)
	if !ok {
		return fmt.Errorf("Dragonfly block.%s is not a struct", name)
	}
	want := []parameter{{Name: "Still", Type: "bool"}, {Name: "Depth", Type: "int"}, {Name: "Falling", Type: "bool"}}
	var got []parameter
	for _, field := range structure.Fields.List {
		identifier, ok := field.Type.(*ast.Ident)
		for _, fieldName := range field.Names {
			if !fieldName.IsExported() {
				continue
			}
			typeName := formatGoExpression(field.Type)
			if ok {
				typeName = identifier.Name
			}
			got = append(got, parameter{Name: fieldName.Name, Type: typeName})
		}
	}
	if !reflect.DeepEqual(got, want) {
		return fmt.Errorf("Dragonfly block.%s fields changed: got %v, want %v", name, got, want)
	}
	return nil
}

func blockFieldVaries(states []world.Block, path []int, fieldType string) bool {
	if len(states) < 2 {
		return false
	}
	value := func(state world.Block) blockFieldValue {
		reflected := reflect.ValueOf(state)
		if reflected.Kind() == reflect.Pointer {
			reflected = reflected.Elem()
		}
		field := reflected.FieldByIndex(path)
		if fieldType == "bool" {
			return blockFieldValue{Bool: field.Bool()}
		}
		return blockFieldValue{Int: int(field.Int())}
	}
	first := value(states[0])
	for _, state := range states[1:] {
		if value(state) != first {
			return true
		}
	}
	return false
}

func blockFieldAtASTPath(name string, path []int, declarations map[string]*ast.TypeSpec) (blockField, error) {
	current := name
	for depth, fieldIndex := range path {
		var structure *ast.StructType
		resolved := map[string]bool{}
		for structure == nil {
			if resolved[current] {
				return blockField{}, fmt.Errorf("Dragonfly block.%s has a recursive underlying type", current)
			}
			resolved[current] = true
			spec := declarations[current]
			if spec == nil {
				return blockField{}, fmt.Errorf("Dragonfly block.%s declaration not found", current)
			}
			switch value := spec.Type.(type) {
			case *ast.StructType:
				structure = value
			case *ast.Ident:
				current = value.Name
			default:
				return blockField{}, fmt.Errorf("Dragonfly block.%s does not have a local struct underlying type", current)
			}
		}
		fields := flattenedASTFields(structure)
		if fieldIndex < 0 || fieldIndex >= len(fields) {
			return blockField{}, fmt.Errorf("Dragonfly block.%s field path %v is outside its AST declaration", name, path)
		}
		field := fields[fieldIndex]
		if depth == len(path)-1 {
			if field.Name == "" {
				return blockField{}, fmt.Errorf("Dragonfly block.%s field path %v ends at an embedding", name, path)
			}
			return blockField{Name: field.Name, Type: formatGoExpression(field.Type)}, nil
		}
		embedded, ok := field.Type.(*ast.Ident)
		if field.Name != "" || !ok {
			return blockField{}, fmt.Errorf("Dragonfly block.%s field path %v does not follow a local embedding", name, path)
		}
		current = embedded.Name
	}
	return blockField{}, fmt.Errorf("Dragonfly block.%s has an empty field path", name)
}

type flattenedASTField struct {
	Name string
	Type ast.Expr
}

func flattenedASTFields(structure *ast.StructType) []flattenedASTField {
	var fields []flattenedASTField
	for _, field := range structure.Fields.List {
		if len(field.Names) == 0 {
			fields = append(fields, flattenedASTField{Type: field.Type})
			continue
		}
		for _, name := range field.Names {
			fields = append(fields, flattenedASTField{Name: name.Name, Type: field.Type})
		}
	}
	return fields
}

func encodeRegisteredBlock(typeName string, value world.Block, registered []world.Block) (encodedBlock, error) {
	identifier, properties := value.EncodeBlock()
	found := false
	for _, candidate := range registered {
		candidateIdentifier, candidateProperties := candidate.EncodeBlock()
		if identifier == candidateIdentifier && reflect.DeepEqual(properties, candidateProperties) {
			found = true
			break
		}
	}
	if !found {
		return encodedBlock{}, fmt.Errorf("block.%s state is not registered", typeName)
	}
	encoded, ok := blockstate.EncodeProperties(properties)
	if !ok {
		return encodedBlock{}, fmt.Errorf("block.%s has unsupported properties", typeName)
	}
	return encodedBlock{Name: typeName, Identifier: identifier, PropertiesNBT: encoded}, nil
}

func syncGeneratedFiles(files []generatedFile, check bool) error {
	for _, file := range files {
		if check {
			current, err := os.ReadFile(file.Path)
			if err != nil || !bytes.Equal(current, file.Content) {
				return fmt.Errorf("%s is stale; run make generate", file.Path)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(file.Path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(file.Path, file.Content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func playerHandlerMethods(path string) ([]method, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	for _, declaration := range file.Decls {
		gen, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "Handler" {
				continue
			}
			interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				return nil, fmt.Errorf("player.Handler is not an interface")
			}
			var methods []method
			for _, field := range interfaceType.Methods.List {
				if len(field.Names) != 1 {
					return nil, fmt.Errorf("player.Handler has an unnamed or multiply named method")
				}
				subscription, supported := supportedPlayerHandlers[field.Names[0].Name]
				if !supported {
					return nil, fmt.Errorf("unsupported player.Handler.%s method", field.Names[0].Name)
				}
				function, ok := field.Type.(*ast.FuncType)
				if !ok {
					return nil, fmt.Errorf("player.Handler.%s is not a method", field.Names[0].Name)
				}
				translated, err := translatePlayerHandlerParameters(field.Names[0].Name, function.Params)
				if err != nil {
					return nil, fmt.Errorf("player.Handler.%s: %w", field.Names[0].Name, err)
				}
				methods = append(methods, method{Name: field.Names[0].Name, Parameters: translated, Subscription: subscription})
			}
			for name := range supportedPlayerHandlers {
				found := false
				for _, method := range methods {
					found = found || method.Name == name
				}
				if !found {
					return nil, fmt.Errorf("Dragonfly player.Handler has no supported %s method", name)
				}
			}
			return methods, nil
		}
	}
	return nil, fmt.Errorf("Dragonfly player.Handler interface not found")
}

func translatePlayerHandlerParameters(methodName string, fields *ast.FieldList) ([]parameter, error) {
	var parameters []parameter
	for _, field := range fields.List {
		if len(field.Names) == 0 {
			return nil, fmt.Errorf("unnamed parameter")
		}
		for _, name := range field.Names {
			if methodName == "HandleChangeWorld" && pointerToSelector(field.Type, "world", "World") {
				switch name.Name {
				case "before":
					parameters = append(parameters, parameter{Name: name.Name, Type: "World?"})
				case "after":
					parameters = append(parameters, parameter{Name: name.Name, Type: "World"})
				default:
					return nil, fmt.Errorf("unsupported world parameter %s", name.Name)
				}
				continue
			}
			typeName, ok := csharpType(field.Type)
			if !ok {
				return nil, fmt.Errorf("unsupported parameter type %T", field.Type)
			}
			parameters = append(parameters, parameter{Name: name.Name, Type: typeName})
		}
	}
	return parameters, nil
}

func pointerToSelector(expression ast.Expr, packageName, typeName string) bool {
	pointer, ok := expression.(*ast.StarExpr)
	if !ok {
		return false
	}
	selector, ok := pointer.X.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != typeName {
		return false
	}
	identifier, ok := selector.X.(*ast.Ident)
	return ok && identifier.Name == packageName
}

func translateParameters(fields *ast.FieldList) ([]parameter, error) {
	var parameters []parameter
	for _, field := range fields.List {
		typeName, ok := csharpType(field.Type)
		if !ok {
			return nil, fmt.Errorf("unsupported parameter type %T", field.Type)
		}
		for _, name := range field.Names {
			parameters = append(parameters, parameter{Name: name.Name, Type: typeName})
		}
	}
	return parameters, nil
}

func csharpType(expression ast.Expr) (string, bool) {
	switch value := expression.(type) {
	case *ast.ArrayType:
		typeName, ok := csharpType(value.Elt)
		if !ok {
			return "", false
		}
		return typeName + "[]", true
	case *ast.Ellipsis:
		typeName, ok := csharpType(value.Elt)
		if !ok {
			return "", false
		}
		return "params " + typeName + "[]", true
	case *ast.StarExpr:
		if nested, ok := value.X.(*ast.StarExpr); ok {
			if selector, ok := nested.X.(*ast.SelectorExpr); ok {
				if packageName, ok := selector.X.(*ast.Ident); ok && packageName.Name == "world" && selector.Sel.Name == "World" {
					return "ref World", true
				}
			}
		}
		typeName, ok := csharpType(value.X)
		if !ok {
			return "", false
		}
		if typeName == "Player.Context" || typeName == "Player" {
			return typeName, true
		}
		return "ref " + typeName, true
	case *ast.Ident:
		typeName, ok := map[string]string{
			"any":     "object?",
			"bool":    "bool",
			"Context": "Player.Context",
			"float64": "double",
			"int":     "int",
			"Player":  "Player",
			"string":  "string",
		}[value.Name]
		return typeName, ok
	case *ast.SelectorExpr:
		packageName, ok := value.X.(*ast.Ident)
		if !ok {
			return "", false
		}
		typeName, ok := map[string]string{
			"cmd.Command":         "Cmd.Command",
			"cube.Face":           "Cube.Face",
			"cube.Pos":            "Cube.Pos",
			"cube.Rotation":       "Rotation",
			"item.Stack":          "Item.Stack",
			"mgl64.Vec3":          "Vector3",
			"net.UDPAddr":         "Net.UDPAddr",
			"session.Diagnostics": "Session.Diagnostics",
			"skin.Skin":           "Skin",
			"time.Duration":       "TimeSpan",
			"world.Block":         "World.Block",
			"world.DamageSource":  "World.DamageSource",
			"world.Entity":        "World.Entity",
			"world.HealingSource": "World.HealingSource",
			"world.World":         "World",
		}[packageName.Name+"."+value.Sel.Name]
		return typeName, ok
	default:
		return "", false
	}
}

func commandInterfaces(directory string) ([]commandInterface, error) {
	packages, err := parser.ParseDir(token.NewFileSet(), directory, nil, 0)
	if err != nil {
		return nil, err
	}
	pkg, ok := packages["cmd"]
	if !ok {
		return nil, fmt.Errorf("Dragonfly cmd package not found")
	}
	found := map[string]commandInterface{}
	for _, file := range pkg.Files {
		for _, declaration := range file.Decls {
			gen, ok := declaration.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gen.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || !selectedCommandInterface(typeSpec.Name.Name) {
					continue
				}
				interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok {
					return nil, fmt.Errorf("cmd.%s is not an interface", typeSpec.Name.Name)
				}
				translated, err := translateCommandInterface(typeSpec.Name.Name, interfaceType)
				if err != nil {
					return nil, err
				}
				found[typeSpec.Name.Name] = translated
			}
		}
	}
	interfaces := make([]commandInterface, 0, len(selectedCommandInterfaces))
	for _, name := range selectedCommandInterfaces {
		definition, ok := found[name]
		if !ok {
			return nil, fmt.Errorf("Dragonfly cmd.%s interface not found", name)
		}
		interfaces = append(interfaces, definition)
	}
	return interfaces, nil
}

func selectedCommandInterface(name string) bool {
	for _, selected := range selectedCommandInterfaces {
		if name == selected {
			return true
		}
	}
	return false
}

func translateCommandInterface(name string, interfaceType *ast.InterfaceType) (commandInterface, error) {
	definition := commandInterface{Name: name}
	for _, field := range interfaceType.Methods.List {
		if len(field.Names) == 0 {
			embedding, ok := field.Type.(*ast.Ident)
			if !ok || !selectedCommandInterface(embedding.Name) {
				return commandInterface{}, fmt.Errorf("cmd.%s has unsupported embedded interface", name)
			}
			definition.Embeddings = append(definition.Embeddings, embedding.Name)
			continue
		}
		if len(field.Names) != 1 {
			return commandInterface{}, fmt.Errorf("cmd.%s has unnamed method", name)
		}
		function, ok := field.Type.(*ast.FuncType)
		if !ok {
			return commandInterface{}, fmt.Errorf("cmd.%s.%s is not a method", name, field.Names[0].Name)
		}
		parameters, err := translateCommandParameters(function.Params)
		if err != nil {
			return commandInterface{}, fmt.Errorf("cmd.%s.%s: %w", name, field.Names[0].Name, err)
		}
		returnType, err := translateCommandResult(function.Results)
		if err != nil {
			return commandInterface{}, fmt.Errorf("cmd.%s.%s: %w", name, field.Names[0].Name, err)
		}
		definition.Methods = append(definition.Methods, commandMethod{
			Name:       field.Names[0].Name,
			Parameters: parameters,
			ReturnType: returnType,
		})
	}
	return definition, nil
}

func translateCommandParameters(fields *ast.FieldList) ([]parameter, error) {
	var parameters []parameter
	if fields == nil {
		return parameters, nil
	}
	for _, field := range fields.List {
		typeName, ok := commandCSharpType(field.Type)
		if !ok {
			return nil, fmt.Errorf("unsupported parameter type %s", formatGoExpression(field.Type))
		}
		for _, name := range field.Names {
			parameters = append(parameters, parameter{Name: name.Name, Type: typeName})
		}
	}
	return parameters, nil
}

func translateCommandResult(fields *ast.FieldList) (string, error) {
	if fields == nil || len(fields.List) == 0 {
		return "void", nil
	}
	if len(fields.List) != 1 || len(fields.List[0].Names) > 1 {
		return "", fmt.Errorf("multiple return values are unsupported")
	}
	typeName, ok := commandCSharpType(fields.List[0].Type)
	if !ok {
		return "", fmt.Errorf("unsupported return type %s", formatGoExpression(fields.List[0].Type))
	}
	return typeName, nil
}

func commandCSharpType(expression ast.Expr) (string, bool) {
	switch value := expression.(type) {
	case *ast.StarExpr:
		typeName, ok := commandCSharpType(value.X)
		if !ok {
			return "", false
		}
		if typeName == "World.Tx" {
			return typeName + "?", true
		}
		return typeName, true
	case *ast.ArrayType:
		if value.Len != nil {
			return "", false
		}
		element, ok := commandCSharpType(value.Elt)
		if !ok {
			return "", false
		}
		return "IReadOnlyList<" + element + ">", true
	case *ast.Ident:
		typeName, ok := map[string]string{
			"bool":   "bool",
			"Output": "Output",
			"Source": "Source",
			"string": "string",
		}[value.Name]
		return typeName, ok
	case *ast.SelectorExpr:
		packageName, ok := value.X.(*ast.Ident)
		if !ok {
			return "", false
		}
		typeName, ok := map[string]string{
			"mgl64.Vec3": "Vector3",
			"world.Tx":   "World.Tx",
		}[packageName.Name+"."+value.Sel.Name]
		return typeName, ok
	default:
		return "", false
	}
}

func formatGoExpression(expression ast.Expr) string {
	switch value := expression.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.SelectorExpr:
		return formatGoExpression(value.X) + "." + value.Sel.Name
	case *ast.StarExpr:
		return "*" + formatGoExpression(value.X)
	case *ast.ArrayType:
		return "[]" + formatGoExpression(value.Elt)
	case *ast.Ellipsis:
		return "..." + formatGoExpression(value.Elt)
	case *ast.MapType:
		return "map[" + formatGoExpression(value.Key) + "]" + formatGoExpression(value.Value)
	case *ast.BinaryExpr:
		return formatGoExpression(value.X) + " " + value.Op.String() + " " + formatGoExpression(value.Y)
	case *ast.FuncType:
		result := "func(" + rawParameterTypes(value.Params) + ")"
		if returns := rawResultTypes(value.Results); returns != "" {
			result += " " + returns
		}
		return result
	default:
		return fmt.Sprintf("%T", expression)
	}
}

func generatePlayerHandler(methods []method) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/handler.go. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\n")
	output.WriteString("namespace Dragonfly;\n\n")
	output.WriteString("public sealed partial class Player\n{\n    public interface Handler\n    {\n")
	for _, method := range methods {
		fmt.Fprintf(&output, "        void %s(%s);\n", method.Name, formatParameters(method.Parameters))
	}
	output.WriteString("    }\n}\n\n")
	output.WriteString("public abstract partial class Plugin\n{\n")
	for _, method := range methods {
		fmt.Fprintf(&output, "    [HandlerSubscription(%dUL)]\n", method.Subscription)
		fmt.Fprintf(&output, "    public virtual void %s(%s) { }\n", method.Name, formatParameters(method.Parameters))
	}
	output.WriteString("}\n")
	return output.Bytes()
}

func generatePlayerTextMethods(methods []method) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\n")
	output.WriteString("using Dragonfly.Native;\n\n")
	output.WriteString("namespace Dragonfly;\n\n")
	output.WriteString("public sealed partial class Player\n{\n")
	for _, method := range methods {
		fmt.Fprintf(&output, "    public void %s(%s) => ", method.Name, formatParameters(method.Parameters))
		switch method.Name {
		case "Message", "SendPopup", "SendTip", "SendJukeboxPopup", "Disconnect":
			kind := map[string]string{
				"Message":          "Message",
				"SendPopup":        "Popup",
				"SendTip":          "Tip",
				"SendJukeboxPopup": "JukeboxPopup",
				"Disconnect":       "Disconnect",
			}[method.Name]
			fmt.Fprintf(&output, "SendText(Abi.PlayerText%s, FormatArguments(%s));\n", kind, method.Parameters[0].Name)
		case "SetNameTag":
			fmt.Fprintf(&output, "SendText(Abi.PlayerTextNameTag, %s);\n", method.Parameters[0].Name)
		default:
			panic("unsupported player text method: " + method.Name)
		}
	}
	output.WriteString("}\n")
	return output.Bytes()
}

func generatePlayerFormMethods(methods []method) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\n\nnamespace Dragonfly;\n\n")
	output.WriteString("public sealed partial class Player : Form.Submitter\n{\n")
	for _, method := range methods {
		fmt.Fprintf(&output, "    public void %s(%s) =>\n", method.Name, formatParameters(method.Parameters))
		switch method.Name {
		case "SendForm":
			fmt.Fprintf(&output, "        PluginBridge.Host.SendPlayerForm(Invocation, Id, %s);\n", method.Parameters[0].Name)
		case "CloseForm":
			output.WriteString("        PluginBridge.Host.ClosePlayerForm(Invocation, Id);\n")
		default:
			panic("unsupported player form method: " + method.Name)
		}
		if method.Name != methods[len(methods)-1].Name {
			output.WriteByte('\n')
		}
	}
	output.WriteString("}\n")
	return output.Bytes()
}

func generateForms(spec formSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/form Go AST. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\nusing System;\nusing System.Collections.Generic;\n\nnamespace Dragonfly;\n\n")
	output.WriteString(`public static partial class Form
{
    public interface Value
    {
        byte[] MarshalJSON();
        void SubmitJSON(byte[]? response, Submitter submitter, World.Tx tx);
    }

    public interface Element
    {
        byte[] MarshalJSON();
    }

    public interface MenuElement
    {
        byte[] MarshalJSON();
    }

    public interface Submittable
    {
        void Submit(Submitter submitter, World.Tx tx);
    }

    public interface MenuSubmittable
    {
        void Submit(Submitter submitter, Button pressed, World.Tx tx);
    }

    public interface ModalSubmittable : MenuSubmittable { }

    public interface Closer
    {
        void Close(Submitter submitter, World.Tx tx);
    }

    public interface Submitter
    {
        void SendForm(Value form);
        void CloseForm();
    }

`)
	for _, element := range spec.Elements {
		generateFormElement(&output, element)
		output.WriteByte('\n')
	}
	output.WriteString(`    public readonly struct Custom : Value
    {
        internal Custom(Submittable submittable, string title) =>
            (Submittable, FormTitle) = (submittable, title);

        internal Submittable Submittable { get; }
        internal string FormTitle { get; }

        public string Title() => FormTitle;
        public IReadOnlyList<Element> Elements() => FormCodec.Elements(this);
        public byte[] MarshalJSON() => FormCodec.Encode(this);
        public void SubmitJSON(byte[]? response, Submitter submitter, World.Tx tx) =>
            FormCodec.Respond(this, submitter, tx, response is null, response ?? Array.Empty<byte>());
    }

    public readonly struct Menu : Value
    {
        internal Menu(MenuSubmittable submittable, string title, string body, MenuElement[] elements) =>
            (Submittable, FormTitle, FormBody, ExtraElements) = (submittable, title, body, elements);

        internal MenuSubmittable Submittable { get; }
        internal string FormTitle { get; }
        internal string FormBody { get; }
        internal MenuElement[] ExtraElements { get; }

        public Menu WithBody(params object?[] body) =>
            new(Submittable, FormTitle, FormCodec.Format(body), ExtraElements);

        public Menu AddButton(Button button) => WithElements(button);
        public Menu AddDivider(Divider divider) => WithElements(divider);
        public Menu AddHeader(Header header) => WithElements(header);
        public Menu AddLabel(Label label) => WithElements(label);

        public Menu WithButtons(params Button[] buttons)
        {
            var elements = new MenuElement[buttons.Length];
            for (var index = 0; index < buttons.Length; index++) elements[index] = buttons[index];
            return WithElements(elements);
        }

        public Menu WithElements(params MenuElement[] elements)
        {
            var combined = new MenuElement[ExtraElements.Length + elements.Length];
            ExtraElements.CopyTo(combined, 0);
            elements.CopyTo(combined, ExtraElements.Length);
            return new Menu(Submittable, FormTitle, FormBody, combined);
        }

        public string Title() => FormTitle;
        public string Body() => FormBody;
        public IReadOnlyList<Button> Buttons() => FormCodec.Buttons(this);
        public IReadOnlyList<MenuElement> Elements() => FormCodec.Elements(this);
        public byte[] MarshalJSON() => FormCodec.Encode(this);
        public void SubmitJSON(byte[]? response, Submitter submitter, World.Tx tx) =>
            FormCodec.Respond(this, submitter, tx, response is null, response ?? Array.Empty<byte>());
    }

    public readonly struct Modal : Value
    {
        internal Modal(ModalSubmittable submittable, string title, string body) =>
            (Submittable, FormTitle, FormBody) = (submittable, title, body);

        internal ModalSubmittable Submittable { get; }
        internal string FormTitle { get; }
        internal string FormBody { get; }

        public Modal WithBody(params object?[] body) =>
            new(Submittable, FormTitle, FormCodec.Format(body));

        public string Title() => FormTitle;
        public string Body() => FormBody;
        public IReadOnlyList<Button> Buttons() => FormCodec.Buttons(this);
        public byte[] MarshalJSON() => FormCodec.Encode(this);
        public void SubmitJSON(byte[]? response, Submitter submitter, World.Tx tx) =>
            FormCodec.Respond(this, submitter, tx, response is null, response ?? Array.Empty<byte>());
    }

    public static Custom New(Submittable submittable, params object?[] title)
    {
        ArgumentNullException.ThrowIfNull(submittable);
        FormCodec.VerifyCustom(submittable);
        return new Custom(submittable, FormCodec.Format(title));
    }

    public static Menu NewMenu(MenuSubmittable submittable, params object?[] title)
    {
        ArgumentNullException.ThrowIfNull(submittable);
        FormCodec.VerifyMenu(submittable);
        return new Menu(submittable, FormCodec.Format(title), string.Empty, Array.Empty<MenuElement>());
    }

    public static Modal NewModal(ModalSubmittable submittable, params object?[] title)
    {
        ArgumentNullException.ThrowIfNull(submittable);
        FormCodec.VerifyModal(submittable);
        return new Modal(submittable, FormCodec.Format(title), string.Empty);
    }

    public static Button YesButton() => NewButton("gui.yes", string.Empty);
    public static Button NoButton() => NewButton("gui.no", string.Empty);
}
`)
	return output.Bytes()
}

func generateFormElement(output *bytes.Buffer, element formElementSpec) {
	fmt.Fprintf(output, "    public struct %s", element.Name)
	var interfaces []string
	if element.Element {
		interfaces = append(interfaces, "Element")
	}
	if element.MenuElement {
		interfaces = append(interfaces, "MenuElement")
	}
	if len(interfaces) != 0 {
		fmt.Fprintf(output, " : %s", strings.Join(interfaces, ", "))
	}
	output.WriteString("\n    {\n")
	for _, field := range element.Fields {
		fmt.Fprintf(output, "        public %s %s;\n", field.Type, field.Name)
	}
	if element.ValueType != "" {
		if element.ValueType == "string" {
			output.WriteString("        private string? _value;\n")
		} else {
			fmt.Fprintf(output, "        private %s _value;\n", element.ValueType)
		}
	}
	output.WriteString("\n        public readonly byte[] MarshalJSON() => ")
	if element.Element {
		output.WriteString("FormCodec.EncodeElement(this);\n")
	} else {
		output.WriteString("FormCodec.EncodeMenuElement(this);\n")
	}
	if element.Tooltip {
		fmt.Fprintf(output, "\n        public readonly %s WithTooltip(string tooltip)\n        {\n", element.Name)
		output.WriteString("            var copy = this;\n            copy.Tooltip = tooltip;\n            return copy;\n        }\n")
	}
	if element.ValueType != "" {
		fmt.Fprintf(output, "\n        public readonly %s Value() => _value", element.ValueType)
		if element.ValueType == "string" {
			output.WriteString(" ?? string.Empty")
		}
		output.WriteString(";\n")
		fmt.Fprintf(output, "\n        internal readonly %s WithValue(%s value)\n        {\n", element.Name, element.ValueType)
		output.WriteString("            var copy = this;\n            copy._value = value;\n            return copy;\n        }\n")
	}
	output.WriteString("    }\n\n")
	if element.Name == "Divider" {
		return
	}
	fmt.Fprintf(output, "    public static %s New%s(%s) => new()\n    {\n", element.Name, element.Name, formatParameters(element.Constructor))
	fields := make([]formFieldSpec, 0, len(element.Fields))
	for _, field := range element.Fields {
		if field.Name != "Tooltip" {
			fields = append(fields, field)
		}
	}
	if len(fields) != len(element.Constructor) {
		panic("form." + element.Name + " constructor no longer maps to exported fields")
	}
	for index, field := range fields {
		fmt.Fprintf(output, "        %s = %s,\n", field.Name, element.Constructor[index].Name)
	}
	output.WriteString("    };\n")
}

func generateCommandInterfaces(interfaces []commandInterface) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/cmd Go AST. DO NOT EDIT.\n#nullable enable\n")
	output.WriteString("namespace Dragonfly;\n\n")
	output.WriteString("public static partial class Cmd\n{\n")
	for index, definition := range interfaces {
		fmt.Fprintf(&output, "    public interface %s", definition.Name)
		if len(definition.Embeddings) != 0 {
			fmt.Fprintf(&output, " : %s", strings.Join(definition.Embeddings, ", "))
		}
		output.WriteString("\n    {\n")
		for _, method := range definition.Methods {
			fmt.Fprintf(&output, "        %s %s(%s);\n", method.ReturnType, method.Name, formatParameters(method.Parameters))
		}
		output.WriteString("    }\n")
		if index != len(interfaces)-1 {
			output.WriteString("\n")
		}
	}
	output.WriteString("}\n")
	return output.Bytes()
}

func generateCube(spec cubeSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/block/cube Go AST. DO NOT EDIT.\n")
	output.WriteString("namespace Dragonfly;\n\n")
	output.WriteString("public static partial class Cube\n{\n")
	output.WriteString("    public enum Face\n    {\n")
	for index, face := range spec.Faces {
		fmt.Fprintf(&output, "        %s = %d,\n", face, index)
	}
	output.WriteString("    }\n\n")
	output.WriteString(`    public readonly record struct Range(int Minimum, int Maximum)
    {
        public int Min() => Minimum;
        public int Max() => Maximum;
        public int Height() => Maximum - Minimum;
    }

    public readonly record struct Pos
    {
        private readonly int _x;
        private readonly int _y;
        private readonly int _z;

        public Pos(int x, int y, int z) => (_x, _y, _z) = (x, y, z);

        public int X() => _x;
        public int Y() => _y;
        public int Z() => _z;
        public bool OutOfBounds(Range range) => _y > range.Max() || _y < range.Min();
        public bool Within(Pos min, Pos max) =>
            _x >= min._x && _x <= max._x &&
            _y >= min._y && _y <= max._y &&
            _z >= min._z && _z <= max._z;
        public Pos Add(Pos other) => new(_x + other._x, _y + other._y, _z + other._z);
        public Pos Sub(Pos other) => new(_x - other._x, _y - other._y, _z - other._z);
        public Vector3 Vec3() => new(_x, _y, _z);
        public Vector3 Vec3Middle() => new(_x + 0.5, _y, _z + 0.5);
        public Vector3 Vec3Centre() => new(_x + 0.5, _y + 0.5, _z + 0.5);

        public Pos Side(Face face) => face switch
        {
            Cube.Face.Up => new(_x, _y + 1, _z),
            Cube.Face.Down => new(_x, _y - 1, _z),
            Cube.Face.North => new(_x, _y, _z - 1),
            Cube.Face.South => new(_x, _y, _z + 1),
            Cube.Face.West => new(_x - 1, _y, _z),
            Cube.Face.East => new(_x + 1, _y, _z),
            _ => this,
        };

        public Face Face(Pos other) => NeighbourFace(other).Face;

        public (Face Face, bool Ok) NeighbourFace(Pos other) => other.Sub(this) switch
        {
            Pos { _x: 0, _y: 1, _z: 0 } => (Cube.Face.Up, true),
            Pos { _x: 0, _y: -1, _z: 0 } => (Cube.Face.Down, true),
            Pos { _x: 0, _y: 0, _z: -1 } => (Cube.Face.North, true),
            Pos { _x: 0, _y: 0, _z: 1 } => (Cube.Face.South, true),
            Pos { _x: -1, _y: 0, _z: 0 } => (Cube.Face.West, true),
            Pos { _x: 1, _y: 0, _z: 0 } => (Cube.Face.East, true),
            _ => (Cube.Face.Up, false),
        };

        public override string ToString() => $"({_x},{_y},{_z})";
    }

    public static Pos PosFromVec3(Vector3 value) => new(
        checked((int)Math.Floor(value.X)),
        checked((int)Math.Floor(value.Y)),
        checked((int)Math.Floor(value.Z)));

    public static Pos Min(Pos first, Pos second) => new(
        Math.Min(first.X(), second.X()),
        Math.Min(first.Y(), second.Y()),
        Math.Min(first.Z(), second.Z()));

    public static Pos Max(Pos first, Pos second) => new(
        Math.Max(first.X(), second.X()),
        Math.Max(first.Y(), second.Y()),
        Math.Max(first.Z(), second.Z()));
}
`)
	return output.Bytes()
}

func generateWorldBlock(setOpts []string, methods []commandMethod) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/world Go AST. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\n")
	for _, method := range methods {
		usesSystem := strings.Contains(method.ReturnType, "TimeSpan")
		for _, parameter := range method.Parameters {
			usesSystem = usesSystem || strings.Contains(parameter.Type, "TimeSpan")
		}
		if usesSystem {
			output.WriteString("using System;\n")
			break
		}
	}
	output.WriteString("using System.Collections.Generic;\n\n")
	output.WriteString("namespace Dragonfly;\n\n")
	output.WriteString("public sealed partial class World\n{\n")
	output.WriteString("    public interface Block { }\n\n")
	output.WriteString("    public static (Block? Block, bool Ok) BlockByName(string name, Dictionary<string, object?>? properties) =>\n")
	output.WriteString("        PluginBridge.Host.BlockByName(name, properties);\n\n")
	output.WriteString("    public interface Biome { }\n\n")
	output.WriteString("    public interface Particle { }\n\n")
	output.WriteString("    public interface Liquid : Block { string LiquidType(); }\n\n")
	output.WriteString("    public sealed class SetOpts\n    {\n")
	for _, field := range setOpts {
		fmt.Fprintf(&output, "        public bool %s;\n", field)
	}
	output.WriteString("    }\n\n")
	output.WriteString("    public partial class Tx\n    {\n")
	for index, method := range methods {
		parameters := formatParameters(method.Parameters)
		if method.Name == "SetBlock" && len(method.Parameters) != 0 {
			last := method.Parameters[len(method.Parameters)-1]
			if last.Name != "opts" || last.Type != "SetOpts?" {
				panic("world.Tx.SetBlock final parameter is not opts *SetOpts")
			}
			parameters = strings.TrimSuffix(parameters, last.Type+" "+last.Name) + last.Type + " " + last.Name + " = null"
		}
		fmt.Fprintf(&output, "        public %s %s(%s) =>\n", method.ReturnType, method.Name, parameters)
		switch method.Name {
		case "World":
			output.WriteString("            PluginBridge.Host.TransactionWorld(Invocation);\n")
		case "Range":
			output.WriteString("            PluginBridge.Host.WorldRange(Invocation);\n")
		case "SetBlock":
			fmt.Fprintf(&output, "            PluginBridge.Host.SetWorldBlock(Invocation, %s, %s, %s);\n",
				method.Parameters[0].Name, method.Parameters[1].Name, method.Parameters[2].Name)
		case "Block":
			fmt.Fprintf(&output, "            PluginBridge.Host.WorldBlock(Invocation, %s);\n", method.Parameters[0].Name)
		case "BlockLoaded":
			fmt.Fprintf(&output, "            PluginBridge.Host.WorldBlockLoaded(Invocation, %s);\n", method.Parameters[0].Name)
		case "BlocksWithin":
			fmt.Fprintf(&output, "            PluginBridge.Host.WorldBlocksWithin(Invocation, %s, %s, %s);\n",
				method.Parameters[0].Name, method.Parameters[1].Name, method.Parameters[2].Name)
		case "Liquid":
			fmt.Fprintf(&output, "            PluginBridge.Host.WorldLiquid(Invocation, %s);\n", method.Parameters[0].Name)
		case "SetLiquid":
			fmt.Fprintf(&output, "            PluginBridge.Host.SetWorldLiquid(Invocation, %s, %s);\n",
				method.Parameters[0].Name, method.Parameters[1].Name)
		case "ScheduleBlockUpdate":
			fmt.Fprintf(&output, "            PluginBridge.Host.ScheduleWorldBlockUpdate(Invocation, %s, %s, %s);\n",
				method.Parameters[0].Name, method.Parameters[1].Name, method.Parameters[2].Name)
		case "HighestLightBlocker":
			fmt.Fprintf(&output, "            PluginBridge.Host.WorldHighestLightBlocker(Invocation, %s, %s);\n",
				method.Parameters[0].Name, method.Parameters[1].Name)
		case "HighestBlock":
			fmt.Fprintf(&output, "            PluginBridge.Host.WorldHighestBlock(Invocation, %s, %s);\n",
				method.Parameters[0].Name, method.Parameters[1].Name)
		case "Light":
			fmt.Fprintf(&output, "            PluginBridge.Host.WorldLight(Invocation, %s);\n", method.Parameters[0].Name)
		case "SkyLight":
			fmt.Fprintf(&output, "            PluginBridge.Host.WorldSkyLight(Invocation, %s);\n", method.Parameters[0].Name)
		case "SetBiome":
			fmt.Fprintf(&output, "            PluginBridge.Host.SetWorldBiome(Invocation, %s, %s);\n",
				method.Parameters[0].Name, method.Parameters[1].Name)
		case "Biome":
			fmt.Fprintf(&output, "            PluginBridge.Host.WorldBiome(Invocation, %s);\n", method.Parameters[0].Name)
		case "Temperature":
			fmt.Fprintf(&output, "            PluginBridge.Host.WorldTemperature(Invocation, %s);\n", method.Parameters[0].Name)
		case "RainingAt":
			fmt.Fprintf(&output, "            PluginBridge.Host.WorldRainingAt(Invocation, %s);\n", method.Parameters[0].Name)
		case "SnowingAt":
			fmt.Fprintf(&output, "            PluginBridge.Host.WorldSnowingAt(Invocation, %s);\n", method.Parameters[0].Name)
		case "ThunderingAt":
			fmt.Fprintf(&output, "            PluginBridge.Host.WorldThunderingAt(Invocation, %s);\n", method.Parameters[0].Name)
		case "Raining":
			output.WriteString("            PluginBridge.Host.WorldRaining(Invocation);\n")
		case "Thundering":
			output.WriteString("            PluginBridge.Host.WorldThundering(Invocation);\n")
		case "CurrentTick":
			output.WriteString("            PluginBridge.Host.WorldCurrentTick(Invocation);\n")
		case "AddParticle":
			fmt.Fprintf(&output, "            PluginBridge.Host.AddWorldParticle(Invocation, %s, %s);\n",
				method.Parameters[0].Name, method.Parameters[1].Name)
		case "PlaySound":
			fmt.Fprintf(&output, "            PluginBridge.Host.PlayWorldSound(Invocation, %s, %s);\n",
				method.Parameters[0].Name, method.Parameters[1].Name)
		case "AddEntity":
			fmt.Fprintf(&output, "            PluginBridge.Host.TransactionAddEntity(Invocation, %s);\n", method.Parameters[0].Name)
		case "AddEntityAt":
			fmt.Fprintf(&output, "            PluginBridge.Host.TransactionAddEntity(Invocation, %s, %s);\n",
				method.Parameters[0].Name, method.Parameters[1].Name)
		case "RemoveEntity":
			fmt.Fprintf(&output, "            PluginBridge.Host.TransactionRemoveEntity(Invocation, %s);\n", method.Parameters[0].Name)
		case "Entities":
			output.WriteString("            PluginBridge.Host.TransactionEntities(Invocation, playersOnly: false);\n")
		case "EntitiesWithin":
			fmt.Fprintf(&output, "            PluginBridge.Host.TransactionEntitiesWithin(Invocation, %s);\n", method.Parameters[0].Name)
		case "Players":
			output.WriteString("            PluginBridge.Host.TransactionEntities(Invocation, playersOnly: true);\n")
		default:
			panic("unsupported world.Tx method: " + method.Name)
		}
		if index != len(methods)-1 {
			output.WriteByte('\n')
		}
	}
	output.WriteString("    }\n}\n")
	return output.Bytes()
}

func generateBlocks(spec blockSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/block Go AST and registry. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\nusing System;\nusing Dragonfly;\n\n")
	output.WriteString("namespace Dragonfly\n{\n    public static partial class Block\n    {\n")
	for _, definition := range spec.Types {
		fmt.Fprintf(&output, "        public readonly record struct %s", definition.Name)
		if len(definition.Fields) != 0 {
			output.WriteByte('(')
			for index, field := range definition.Fields {
				if index != 0 {
					output.WriteString(", ")
				}
				fmt.Fprintf(&output, "%s %s = %s", field.Type, field.Name, csharpBlockFieldValue(field, blockFieldValue{}))
			}
			output.WriteByte(')')
		}
		output.WriteString(" : World.Block;\n")
	}
	for _, liquid := range spec.Liquids {
		fmt.Fprintf(&output, "        public readonly record struct %s(bool Still, int Depth, bool Falling) : World.Liquid\n        {\n", liquid.Name)
		fmt.Fprintf(&output, "            public string LiquidType() => %s;\n        }\n", strconv.Quote(liquid.LiquidType))
	}
	output.WriteString("    }\n\n")
	output.WriteString("    internal static class BlockCodec\n    {\n")
	var states []encodedBlock
	for _, definition := range spec.Types {
		for _, state := range definition.States {
			states = append(states, state.encodedBlock)
		}
	}
	for _, liquid := range spec.Liquids {
		for _, state := range liquid.States {
			states = append(states, state.encodedBlock)
		}
	}
	for index, state := range states {
		fmt.Fprintf(&output, "        private static readonly byte[] State%d = %s;\n", index, csharpBytes(state.PropertiesNBT))
	}
	output.WriteString("\n        internal static bool TryEncode(World.Block block, out string identifier, out byte[] properties)\n        {\n")
	output.WriteString("            switch (block)\n            {\n")
	stateIndex := 0
	for _, definition := range spec.Types {
		for _, state := range definition.States {
			fmt.Fprintf(&output, "                case Block.%s", definition.Name)
			if len(definition.Fields) == 0 {
				output.WriteString(" _")
			} else {
				output.WriteString(" { ")
				for index, field := range definition.Fields {
					if index != 0 {
						output.WriteString(", ")
					}
					fmt.Fprintf(&output, "%s: %s", field.Name, csharpBlockFieldValue(field, state.Values[index]))
				}
				output.WriteString(" }")
			}
			fmt.Fprintf(&output, ":\n                    identifier = %s; properties = State%d; return true;\n", strconv.Quote(state.Identifier), stateIndex)
			stateIndex++
		}
	}
	liquidOffset := stateIndex
	for _, liquid := range spec.Liquids {
		for _, state := range liquid.States {
			fmt.Fprintf(&output, "                case Block.%s { Still: %t, Depth: %d, Falling: %t }:\n                    identifier = %s; properties = State%d; return true;\n",
				liquid.Name, state.Still, state.Depth, state.Falling, strconv.Quote(state.Identifier), liquidOffset)
			liquidOffset++
		}
	}
	output.WriteString("                case EncodedLiquid liquidEncoded:\n                    identifier = liquidEncoded.Identifier; properties = liquidEncoded.Properties; return true;\n")
	output.WriteString("                case EncodedBlock encoded:\n                    identifier = encoded.Identifier; properties = encoded.Properties; return true;\n")
	output.WriteString("                default:\n                    identifier = string.Empty; properties = Array.Empty<byte>(); return false;\n            }\n        }\n\n")
	output.WriteString("        internal static World.Block Decode(string identifier, ReadOnlySpan<byte> properties)\n        {\n")
	blockCases, liquidCases := blockDecodeCases(spec)
	writeBlockDecodeCases(&output, append(blockCases, liquidCases...))
	output.WriteString("            return new EncodedBlock(identifier, properties.ToArray());\n        }\n\n")
	output.WriteString("        internal static World.Liquid DecodeLiquid(string identifier, ReadOnlySpan<byte> properties)\n        {\n")
	writeBlockDecodeCases(&output, liquidCases)
	output.WriteString("            return new EncodedLiquid(identifier, properties.ToArray());\n        }\n\n")
	output.WriteString("        private sealed record EncodedBlock(string Identifier, byte[] Properties) : World.Block;\n")
	output.WriteString("        private sealed record EncodedLiquid(string Identifier, byte[] Properties) : World.Liquid\n        {\n            public string LiquidType() => throw new InvalidOperationException(\"Opaque liquid type was not transported by the host.\");\n        }\n")
	output.WriteString("    }\n}\n")
	return output.Bytes()
}

func csharpBlockFieldValue(field blockField, value blockFieldValue) string {
	if field.Type == "bool" {
		return strconv.FormatBool(value.Bool)
	}
	return strconv.Itoa(value.Int)
}

func blockDecodeCases(spec blockSpec) (blocks, liquids []blockDecodeCase) {
	stateIndex := 0
	for _, definition := range spec.Types {
		for _, state := range definition.States {
			var constructor strings.Builder
			fmt.Fprintf(&constructor, "new Block.%s(", definition.Name)
			for index, field := range definition.Fields {
				if index != 0 {
					constructor.WriteString(", ")
				}
				constructor.WriteString(csharpBlockFieldValue(field, state.Values[index]))
			}
			constructor.WriteByte(')')
			blocks = append(blocks, blockDecodeCase{Identifier: state.Identifier, State: stateIndex, Constructor: constructor.String()})
			stateIndex++
		}
	}
	for _, liquid := range spec.Liquids {
		for _, state := range liquid.States {
			liquids = append(liquids, blockDecodeCase{
				Identifier: state.Identifier,
				State:      stateIndex,
				Constructor: fmt.Sprintf("new Block.%s(%t, %d, %t)",
					liquid.Name, state.Still, state.Depth, state.Falling),
			})
			stateIndex++
		}
	}
	return blocks, liquids
}

func writeBlockDecodeCases(output *bytes.Buffer, cases []blockDecodeCase) {
	grouped := map[string][]blockDecodeCase{}
	for _, state := range cases {
		grouped[state.Identifier] = append(grouped[state.Identifier], state)
	}
	identifiers := make([]string, 0, len(grouped))
	for identifier := range grouped {
		identifiers = append(identifiers, identifier)
	}
	sort.Strings(identifiers)
	output.WriteString("            switch (identifier)\n            {\n")
	for _, identifier := range identifiers {
		fmt.Fprintf(output, "                case %s:\n", strconv.Quote(identifier))
		for _, state := range grouped[identifier] {
			fmt.Fprintf(output, "                    if (properties.SequenceEqual(State%d)) return %s;\n", state.State, state.Constructor)
		}
		output.WriteString("                    break;\n")
	}
	output.WriteString("            }\n")
}

func generateBiomes(biomes []encodedBiome) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/world/biome Go AST and registry. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\n\n")
	output.WriteString("namespace Dragonfly\n{\n    public static partial class Biome\n    {\n")
	for _, biome := range biomes {
		fmt.Fprintf(&output, "        public readonly record struct %s : World.Biome;\n", biome.Name)
	}
	output.WriteString("    }\n\n")
	output.WriteString("    internal static class BiomeCodec\n    {\n")
	output.WriteString("        internal static bool TryEncode(World.Biome biome, out int id)\n        {\n")
	output.WriteString("            switch (biome)\n            {\n")
	for _, biome := range biomes {
		fmt.Fprintf(&output, "                case Biome.%s _:\n                    id = %d; return true;\n", biome.Name, biome.ID)
	}
	output.WriteString("                case EncodedBiome encoded:\n                    id = encoded.Id; return true;\n")
	output.WriteString("                default:\n                    id = 0; return false;\n            }\n        }\n\n")
	output.WriteString("        internal static World.Biome Decode(int id)\n        {\n")
	for _, biome := range biomes {
		fmt.Fprintf(&output, "            if (id == %d) return new Biome.%s();\n", biome.ID, biome.Name)
	}
	output.WriteString("            return new EncodedBiome(id);\n        }\n\n")
	output.WriteString("        private sealed record EncodedBiome(int Id) : World.Biome;\n")
	output.WriteString("    }\n}\n")
	return output.Bytes()
}

func generateItems(spec itemSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/item Go AST and live registry. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\nusing System;\nusing System.Collections.Generic;\nusing System.Text;\n\nnamespace Dragonfly\n{\n")
	output.WriteString("    public static partial class Item\n    {\n")
	fmt.Fprintf(&output, "        public readonly record struct ToolTier(%s);\n\n", formatParameters(spec.ToolTierFields))
	for _, tier := range spec.ToolTiers {
		fmt.Fprintf(&output, "        public static readonly ToolTier %s = new(%s);\n", tier.Variable, csharpToolTier(tier.Value))
	}
	output.WriteByte('\n')
	generateEnchantments(&output, spec.Enchantments)
	for _, valueType := range spec.ValueTypes {
		if valueType.Container != "Item" {
			continue
		}
		generateItemValueType(&output, valueType, "        ")
	}
	generateArmourTypes(&output, spec.Armour)
	if itemTypeByName(spec.Types, "Firework") != nil {
		generateFireworkExplosionType(&output)
	}
	for _, definition := range spec.Types {
		if definition.Bucket {
			generateBucketType(&output, spec.Bucket)
			continue
		}
		if definition.Armour {
			generateArmourPieceType(&output, *armourPieceByName(spec.Armour.Pieces, definition.Name), spec.Armour)
			continue
		}
		if definition.NBT {
			generateNBTItemType(&output, definition.Name, spec)
			continue
		}
		if len(definition.Fields) == 0 {
			if material := armourTrimMaterialByName(spec.Armour.TrimMaterials, definition.Name); material != nil {
				fmt.Fprintf(&output, "        public readonly record struct %s : World.Item, ArmourTrimMaterial\n        {\n", definition.Name)
				fmt.Fprintf(&output, "            public string TrimMaterial() => %s;\n", strconv.Quote(material.Material))
				fmt.Fprintf(&output, "            public string MaterialColour() => %s;\n        }\n", strconv.Quote(material.MaterialColour))
			} else if itemTypeFuel(definition) {
				fmt.Fprintf(&output, "        public readonly record struct %s : World.Item, Fuel\n        {\n", definition.Name)
				output.WriteString("            public FuelInfo FuelInfo() => ItemCapabilities.FuelInfo(this);\n        }\n")
			} else {
				fmt.Fprintf(&output, "        public readonly record struct %s : World.Item;\n", definition.Name)
			}
			continue
		}
		parameters := make([]parameter, len(definition.Fields))
		for index, field := range definition.Fields {
			typeName := "bool"
			switch field.Kind {
			case itemFieldToolTier:
				typeName = "ToolTier"
			case itemFieldValue:
				typeName = findItemValueType(spec.ValueTypes, field.ValueType).CSharpType
			}
			parameters[index] = parameter{Name: field.Name, Type: typeName}
		}
		if itemTypeFuel(definition) {
			fmt.Fprintf(&output, "        public readonly record struct %s(%s) : World.Item, Fuel\n        {\n", definition.Name, formatParameters(parameters))
			output.WriteString("            public FuelInfo FuelInfo() => ItemCapabilities.FuelInfo(this);\n        }\n")
		} else {
			fmt.Fprintf(&output, "        public readonly record struct %s(%s) : World.Item;\n", definition.Name, formatParameters(parameters))
		}
	}
	output.WriteString("    }\n\n")
	for _, container := range []string{"Potion", "Sound"} {
		fmt.Fprintf(&output, "    public static partial class %s\n    {\n", container)
		for _, valueType := range spec.ValueTypes {
			if valueType.Container == container {
				generateItemValueType(&output, valueType, "        ")
			}
		}
		output.WriteString("    }\n\n")
	}
	output.WriteString("    internal static class ItemCodec\n    {\n")
	output.WriteString("        internal static bool TryEncode(World.Item item, out string identifier, out int metadata)\n        {\n")
	output.WriteString("            switch (item)\n            {\n")
	for _, definition := range spec.Types {
		for _, state := range definition.States {
			fmt.Fprintf(&output, "                case Item.%s", definition.Name)
			switch {
			case definition.Bucket:
				output.WriteString(strings.TrimPrefix(csharpBucketPattern(state.Bucket), "Item.Bucket"))
				output.WriteString(":\n")
			case definition.Armour:
				fmt.Fprintf(&output, " value when value.Tier is Item.%s:\n", spec.Armour.Tiers[state.ArmourTier].Name)
			case definition.Name == "FireworkStar":
				valueType := findItemValueType(spec.ValueTypes, "Colour")
				fmt.Fprintf(&output, " value when value.FireworkExplosion.Colour == %s:\n", itemValueFactory(*valueType, state.Values[0]))
			case len(definition.Fields) == 0:
				output.WriteString(" _:\n")
			case definition.Fields[0].Kind == itemFieldToolTier:
				fmt.Fprintf(&output, " value when value.%s == Item.%s:\n", definition.Fields[0].Name, spec.ToolTiers[state.ToolTier].Variable)
			case definition.Fields[0].Kind == itemFieldValue:
				valueType := findItemValueType(spec.ValueTypes, definition.Fields[0].ValueType)
				fmt.Fprintf(&output, " value when value.%s == %s:\n", definition.Fields[0].Name, itemValueFactory(*valueType, state.Values[0]))
			default:
				output.WriteString(" { ")
				for index, field := range definition.Fields {
					if index != 0 {
						output.WriteString(", ")
					}
					fmt.Fprintf(&output, "%s: %t", field.Name, state.Bools[index])
				}
				output.WriteString(" }:\n")
			}
			fmt.Fprintf(&output, "                    identifier = %s; metadata = %d; return true;\n", strconv.Quote(state.Identifier), state.Metadata)
		}
	}
	if spec.Bucket.Present {
		output.WriteString("                case Item.Bucket value when value.Content.RawLiquid is not null:\n")
		output.WriteString("                    identifier = \"minecraft:\" + value.Content.String() + \"_bucket\"; metadata = 0; return true;\n")
	}
	output.WriteString("                case EncodedItem encoded:\n                    identifier = encoded.Identifier; metadata = encoded.Metadata; return true;\n")
	output.WriteString("                default:\n                    identifier = string.Empty; metadata = 0; return false;\n            }\n        }\n\n")
	output.WriteString("        internal static World.Item Decode(string identifier, int metadata)\n        {\n")
	for _, definition := range spec.Types {
		for _, state := range definition.States {
			fmt.Fprintf(&output, "            if (identifier == %s && metadata == %d) return new Item.%s(", strconv.Quote(state.Identifier), state.Metadata, definition.Name)
			switch {
			case definition.Bucket:
				output.WriteString(csharpBucketContent(state.Bucket))
			case definition.Armour:
				fmt.Fprintf(&output, "new Item.%s()", spec.Armour.Tiers[state.ArmourTier].Name)
			case definition.Name == "FireworkStar":
				valueType := findItemValueType(spec.ValueTypes, "Colour")
				fmt.Fprintf(&output, "new Item.FireworkExplosion { Colour = %s }", itemValueFactory(*valueType, state.Values[0]))
			case len(definition.Fields) == 0:
			case definition.Fields[0].Kind == itemFieldToolTier:
				output.WriteString("Item." + spec.ToolTiers[state.ToolTier].Variable)
			case definition.Fields[0].Kind == itemFieldValue:
				valueType := findItemValueType(spec.ValueTypes, definition.Fields[0].ValueType)
				output.WriteString(itemValueFactory(*valueType, state.Values[0]))
			default:
				for index, value := range state.Bools {
					if index != 0 {
						output.WriteString(", ")
					}
					output.WriteString(strconv.FormatBool(value))
				}
			}
			output.WriteString(");\n")
		}
	}
	output.WriteString("            return new EncodedItem(identifier, metadata, []);\n        }\n\n")
	fmt.Fprintf(&output, "        internal static bool IsAir(World.Item item) =>\n            TryEncode(item, out var identifier, out _) && identifier == %s;\n\n", strconv.Quote(spec.AirIdentifier))
	output.WriteString("        internal static bool TryRawNbt(World.Item? item, out byte[] nbt)\n        {\n            if (item is EncodedItem encoded)\n            {\n                nbt = encoded.Nbt;\n                return true;\n            }\n            nbt = [];\n            return false;\n        }\n\n")
	output.WriteString("        internal static bool TryWithRawNbt(World.Item item, ReadOnlySpan<byte> nbt, out World.Item result)\n        {\n            if (item is EncodedItem encoded)\n            {\n                result = encoded with { Nbt = nbt.ToArray() };\n                return true;\n            }\n            result = item;\n            return false;\n        }\n\n")
	output.WriteString("        private sealed record EncodedItem(string Identifier, int Metadata, byte[] Nbt) : World.Item;\n")
	output.WriteString("    }\n\n")
	generateArmourCodec(&output, spec)
	generateItemCapabilities(&output, spec)
	output.WriteString("}\n")
	return output.Bytes()
}

func generateEnchantments(output *bytes.Buffer, enchantments []enchantmentSpec) {
	output.WriteString(`        public interface EnchantmentType
        {
            string Name();
            int MaxLevel();
        }

        public readonly record struct Enchantment
        {
            private readonly EnchantmentType? _type;
            private readonly int _level;

            internal Enchantment(EnchantmentType type, int level)
            {
                _type = type;
                _level = level;
            }

            public int Level() => _level;
            public EnchantmentType? Type() => _type;
        }

        public static Enchantment NewEnchantment(EnchantmentType type, int level)
        {
            ArgumentNullException.ThrowIfNull(type);
            if (level < 1) throw new ArgumentOutOfRangeException(nameof(level));
            return new Enchantment(type, level);
        }

`)
	for _, enchantment := range enchantments {
		fmt.Fprintf(output, "        public static readonly EnchantmentType %s = new BuiltinEnchantmentType(%d, %s, %d, %dUL);\n",
			enchantment.Name, enchantment.ID, strconv.Quote(enchantment.DisplayName), enchantment.MaxLevel, enchantment.CompatibleEnchantments)
	}
	output.WriteString(`
        public static (EnchantmentType? Type, bool Ok) EnchantmentByID(int id)
        {
            var enchantment = EnchantmentTypeByID(id);
            return (enchantment, enchantment is not null);
        }

        public static (int ID, bool Ok) EnchantmentID(EnchantmentType type)
        {
            ArgumentNullException.ThrowIfNull(type);
            return TryEnchantmentID(type, out var id) ? (id, true) : (0, false);
        }

        public static IReadOnlyList<EnchantmentType> Enchantments() => new EnchantmentType[]
        {
`)
	for _, enchantment := range enchantments {
		fmt.Fprintf(output, "            %s,\n", enchantment.Name)
	}
	output.WriteString(`        };

        internal static bool TryEnchantmentID(EnchantmentType? type, out int id)
        {
            if (type is BuiltinEnchantmentType builtin)
            {
                id = builtin.ID;
                return true;
            }
            id = 0;
            return false;
        }

        internal static EnchantmentType? EnchantmentTypeByID(int id) => id switch
        {
`)
	for _, enchantment := range enchantments {
		fmt.Fprintf(output, "            %d => %s,\n", enchantment.ID, enchantment.Name)
	}
	output.WriteString(`            _ => null,
        };

        internal static bool EnchantmentsCompatible(EnchantmentType adding, EnchantmentType existing)
        {
            if (!TryEnchantmentID(adding, out var addingID) ||
                !TryEnchantmentID(existing, out var existingID) ||
                EnchantmentTypeByID(addingID) is not BuiltinEnchantmentType builtin)
                return false;
            return (builtin.CompatibleEnchantments & (1UL << existingID)) != 0;
        }

        private sealed record BuiltinEnchantmentType(
            int ID,
            string DisplayName,
            int MaximumLevel,
            ulong CompatibleEnchantments) : EnchantmentType
        {
            public string Name() => DisplayName;
            public int MaxLevel() => MaximumLevel;
        }

`)
}

func generateBucketType(output *bytes.Buffer, spec bucketSpec) {
	if !spec.Present {
		return
	}
	residue := "default"
	if spec.FuelResidueIdentifier != "" && spec.FuelResidueCount > 0 {
		if spec.FuelResidueIdentifier != "minecraft:bucket" || spec.FuelResidueMetadata != 0 {
			panic("bucket fuel residue is not an empty bucket")
		}
		residue = fmt.Sprintf("NewStack(new Bucket(), %d)", spec.FuelResidueCount)
	}
	fmt.Fprintf(output, `        public readonly struct BucketContent
        {
            private readonly World.Liquid? _liquid;
            private readonly bool _milk;

            internal BucketContent(World.Liquid? liquid, bool milk)
            {
                _liquid = liquid;
                _milk = milk;
            }

            internal World.Liquid? RawLiquid => _liquid;
            internal bool Milk => _milk;

            public (World.Liquid? Liquid, bool Ok) Liquid() => (_liquid, _liquid is not null);

            public string String() => _milk ? "milk" : _liquid?.LiquidType() ?? string.Empty;

            public string LiquidType() => _liquid is null ? "milk" : String();
            public override string ToString() => String();
        }

        public static BucketContent LiquidBucketContent(World.Liquid liquid)
        {
            ArgumentNullException.ThrowIfNull(liquid);
            return new BucketContent(liquid, false);
        }

        public static BucketContent MilkBucketContent() => new(null, true);

        public readonly record struct Bucket(BucketContent Content = default) : World.Item, MaxCounter, Fuel
        {
            public int MaxCount() => Empty() ? %d : %d;
            public bool AlwaysConsumable() => Content.Milk;
            public bool CanConsume() => Content.Milk;
            public TimeSpan ConsumeDuration() => TimeSpan.FromTicks(%d);
            public bool Empty() => Content.RawLiquid is null && !Content.Milk;
            public FuelInfo FuelInfo() => Content.RawLiquid?.LiquidType() == "lava"
                ? new FuelInfo(TimeSpan.FromTicks(%d), %s)
                : default;
        }

`, spec.EmptyMaxCount, spec.FullMaxCount, csharpDurationTicks(spec.ConsumeDuration), csharpDurationTicks(spec.FuelDuration), residue)
}

func csharpBucketPattern(content bucketContentKind) string {
	switch content {
	case bucketEmpty:
		return "Item.Bucket value when value.Empty()"
	case bucketWater:
		return "Item.Bucket value when value.Content.RawLiquid is Block.Water"
	case bucketLava:
		return "Item.Bucket value when value.Content.RawLiquid is Block.Lava"
	case bucketMilk:
		return "Item.Bucket value when value.Content.Milk"
	default:
		panic("unsupported bucket content")
	}
}

func csharpBucketContent(content bucketContentKind) string {
	switch content {
	case bucketEmpty:
		return ""
	case bucketWater:
		return "Item.LiquidBucketContent(new Block.Water(false, 0, false))"
	case bucketLava:
		return "Item.LiquidBucketContent(new Block.Lava(false, 0, false))"
	case bucketMilk:
		return "Item.MilkBucketContent()"
	default:
		panic("unsupported bucket content")
	}
}

func generateArmourTypes(output *bytes.Buffer, spec armourSpec) {
	if len(spec.Tiers) == 0 {
		return
	}
	output.WriteString(`        public interface Armour
        {
            double DefencePoints();
            double Toughness();
            double KnockBackResistance();
        }

        public interface ArmourTier
        {
            double BaseDurability();
            double Toughness();
            double KnockBackResistance();
            int EnchantmentValue();
            string Name();
        }

        public interface HelmetType : Armour { bool Helmet(); }
        public interface ChestplateType : Armour { bool Chestplate(); }
        public interface LeggingsType : Armour { bool Leggings(); }
        public interface BootsType : Armour { bool Boots(); }

        public interface ArmourTrimMaterial
        {
            string TrimMaterial();
            string MaterialColour();
        }

        public interface Trimmable { World.Item WithTrim(ArmourTrim trim); }
        public interface MaxCounter { int MaxCount(); }
        public interface Enchantable { int EnchantmentValue(); }
        public interface Durable { DurabilityInfo DurabilityInfo(); }
        public interface Repairable : Durable { bool RepairableBy(Stack stack); }
        public interface Smeltable { SmeltInfo SmeltInfo(); }
        public interface Fuel { FuelInfo FuelInfo(); }

        public readonly record struct ArmourTrim(SmithingTemplateType Template, ArmourTrimMaterial? Material)
        {
            public bool Zero() => Material is null || Template == TemplateNetheriteUpgrade();
        }

        public readonly record struct DurabilityInfo(
            int MaxDurability,
            Func<Stack>? BrokenItem = null,
            int AttackDurability = 0,
            int BreakDurability = 0,
            bool Persistent = false);

        public readonly record struct SmeltInfo(
            Stack Product = default,
            double Experience = 0d,
            bool Food = false,
            bool Ores = false);

        public readonly record struct FuelInfo(TimeSpan Duration = default, Stack Residue = default)
        {
            public FuelInfo WithResidue(Stack residue) => this with { Residue = residue };
        }

`)
	for _, tier := range spec.Tiers {
		if tier.Colour {
			fmt.Fprintf(output, "        public readonly record struct %s(global::Dragonfly.Color.RGBA Colour = default) : ArmourTier\n", tier.Name)
		} else {
			fmt.Fprintf(output, "        public readonly record struct %s : ArmourTier\n", tier.Name)
		}
		output.WriteString("        {\n")
		fmt.Fprintf(output, "            public double BaseDurability() => %s;\n", csharpDouble(tier.BaseDurability))
		fmt.Fprintf(output, "            public double Toughness() => %s;\n", csharpDouble(tier.Toughness))
		fmt.Fprintf(output, "            public double KnockBackResistance() => %s;\n", csharpDouble(tier.KnockBackResistance))
		fmt.Fprintf(output, "            public int EnchantmentValue() => %d;\n", tier.EnchantmentValue)
		fmt.Fprintf(output, "            public string Name() => %s;\n        }\n\n", strconv.Quote(tier.IdentifierName))
	}
	output.WriteString("        public static ArmourTier[] ArmourTiers() =>\n        [\n")
	for _, tier := range spec.Tiers {
		fmt.Fprintf(output, "            new %s(),\n", tier.Name)
	}
	output.WriteString("        ];\n\n        public static World.Item[] ArmourTrimMaterials() =>\n        [\n")
	for _, material := range spec.TrimMaterials {
		fmt.Fprintf(output, "            new %s(),\n", material.ItemName)
	}
	output.WriteString("        ];\n\n")
}

func generateArmourPieceType(output *bytes.Buffer, piece armourPieceSpec, spec armourSpec) {
	fmt.Fprintf(output, "        public readonly record struct %s(ArmourTier Tier, ArmourTrim Trim = default) : World.Item, %sType, Trimmable, MaxCounter, Enchantable, Repairable, Smeltable\n        {\n", piece.Name, piece.SlotMethod)
	output.WriteString("            public int MaxCount() => 1;\n")
	output.WriteString("            public double DefencePoints() => Tier.Name() switch\n            {\n")
	for tierIndex, tier := range spec.Tiers {
		fmt.Fprintf(output, "                %s => %s,\n", strconv.Quote(tier.IdentifierName), csharpDouble(piece.DefencePoints[tierIndex]))
	}
	fmt.Fprintf(output, "                _ => throw new InvalidOperationException(%s),\n            };\n", strconv.Quote("invalid "+strings.ToLower(piece.Name)+" tier"))
	output.WriteString("            public double Toughness() => Tier.Toughness();\n")
	output.WriteString("            public double KnockBackResistance() => Tier.KnockBackResistance();\n")
	output.WriteString("            public int EnchantmentValue() => Tier.EnchantmentValue();\n")
	if piece.DurabilityDivisor == 0 {
		output.WriteString("            public DurabilityInfo DurabilityInfo() => new((int)Tier.BaseDurability(), static () => default);\n")
	} else {
		fmt.Fprintf(output, "            public DurabilityInfo DurabilityInfo()\n            {\n                var value = Tier.BaseDurability();\n                return new((int)(value + value / %s), static () => default);\n            }\n", csharpDouble(piece.DurabilityDivisor))
	}
	output.WriteString("            public SmeltInfo SmeltInfo() => Tier switch\n            {\n")
	for tierIndex, tier := range spec.Tiers {
		smelt := piece.Smelts[tierIndex]
		if smelt.Product == "" {
			continue
		}
		fmt.Fprintf(output, "                %s => new(NewStack(new %s(), %d), %s, %s, %s),\n",
			tier.Name, smelt.Product, smelt.Count, csharpDouble(smelt.Experience), strconv.FormatBool(smelt.Food), strconv.FormatBool(smelt.Ores))
	}
	output.WriteString("                _ => default,\n            };\n")
	output.WriteString("            public bool RepairableBy(Stack stack) => Tier switch\n            {\n")
	for tierIndex, tier := range spec.Tiers {
		fmt.Fprintf(output, "                %s => stack.Item() is %s,\n", tier.Name, piece.RepairItems[tierIndex])
	}
	output.WriteString("                _ => false,\n            };\n")
	fmt.Fprintf(output, "            bool %sType.%s() => true;\n", piece.SlotMethod, piece.SlotMethod)
	fmt.Fprintf(output, "            public World.Item WithTrim(ArmourTrim trim) => new %s(Tier, trim);\n", piece.Name)
	output.WriteString("        }\n\n")
}

func generateArmourCodec(output *bytes.Buffer, spec itemSpec) {
	if len(spec.Armour.Tiers) == 0 {
		return
	}
	output.WriteString("    internal static class ArmourCodec\n    {\n")
	output.WriteString("        internal static bool TryTrimMaterial(string name, out Item.ArmourTrimMaterial? material)\n        {\n            switch (name)\n            {\n")
	for _, material := range spec.Armour.TrimMaterials {
		fmt.Fprintf(output, "                case %s: material = new Item.%s(); return true;\n", strconv.Quote(material.Material), material.ItemName)
	}
	output.WriteString("                default: material = null; return false;\n            }\n        }\n\n")
	templates := findItemValueType(spec.ValueTypes, "SmithingTemplateType")
	output.WriteString("        internal static bool TryTemplate(string name, out Item.SmithingTemplateType template)\n        {\n            switch (name)\n            {\n")
	for index, value := range templates.Values {
		stringer, ok := value.(interface{ String() string })
		if !ok {
			panic(fmt.Sprintf("item.SmithingTemplateType value %T has no String method", value))
		}
		fmt.Fprintf(output, "                case %s: template = %s; return true;\n", strconv.Quote(stringer.String()), itemValueFactory(*templates, index))
	}
	output.WriteString("                default: template = default; return false;\n            }\n        }\n    }\n\n")
}

func armourPieceByName(pieces []armourPieceSpec, name string) *armourPieceSpec {
	for index := range pieces {
		if pieces[index].Name == name {
			return &pieces[index]
		}
	}
	return nil
}

func armourTrimMaterialByName(materials []armourTrimMaterialSpec, name string) *armourTrimMaterialSpec {
	for index := range materials {
		if materials[index].ItemName == name {
			return &materials[index]
		}
	}
	return nil
}

func itemTypeByName(types []itemTypeSpec, name string) *itemTypeSpec {
	for index := range types {
		if types[index].Name == name {
			return &types[index]
		}
	}
	return nil
}

func generateFireworkExplosionType(output *bytes.Buffer) {
	output.WriteString(`        public readonly record struct FireworkExplosion
        {
            public FireworkShape Shape { get; init; }
            public Colour Colour { get; init; }
            public Colour Fade { get; init; }
            public bool Fades { get; init; }
            public bool Twinkle { get; init; }
            public bool Trail { get; init; }
        }

`)
}

func generateNBTItemType(output *bytes.Buffer, name string, spec itemSpec) {
	if name == "Crossbow" {
		fmt.Fprintf(output, `        public readonly record struct Crossbow(Stack Item = default) : World.Item, MaxCounter, Durable, Fuel, Enchantable
        {
            public int MaxCount() => %d;
            public DurabilityInfo DurabilityInfo() => new(%d, static () => default);
            public FuelInfo FuelInfo() => new(TimeSpan.FromTicks(%d));
            public int EnchantmentValue() => %d;
        }

`, spec.Crossbow.MaxCount, spec.Crossbow.MaxDurability, csharpDurationTicks(spec.Crossbow.FuelDuration), spec.Crossbow.EnchantmentValue)
		return
	}
	if name == "Firework" {
		output.WriteString(`        public readonly struct Firework : World.Item
        {
            public Firework(TimeSpan Duration = default, params FireworkExplosion[] Explosions)
            {
                ArgumentNullException.ThrowIfNull(Explosions);
                this.Duration = Duration;
                _explosions = Explosions;
            }

            private readonly FireworkExplosion[]? _explosions;
            public TimeSpan Duration { get; }
            public FireworkExplosion[] Explosions => _explosions ?? [];
            public TimeSpan RandomisedDuration() => Duration + TimeSpan.FromTicks(Random.Shared.NextInt64(6_000_000));
            public bool OffHand() => true;
        }

`)
		return
	}
	if name == "FireworkStar" {
		output.WriteString(`        public readonly record struct FireworkStar(FireworkExplosion FireworkExplosion) : World.Item;

`)
		return
	}
	if name == "BookAndQuill" {
		output.WriteString(`        public readonly struct BookAndQuill : World.Item
        {
            public BookAndQuill(params string[] pages)
            {
                ArgumentNullException.ThrowIfNull(pages);
                _pages = pages;
            }

            private readonly string[]? _pages;
            public string[] Pages => _pages ?? [];
            public int TotalPages() => Pages.Length;

            public (string Page, bool Ok) Page(int page) => page >= 0 && page < TotalPages()
                ? (Pages[page], true)
                : (string.Empty, false);

            public BookAndQuill DeletePage(int page)
            {
                if (page is < 0 or >= 50) throw new ArgumentOutOfRangeException(nameof(page));
                if (page >= TotalPages()) throw new InvalidOperationException("cannot delete nonexistent page");
                var pages = new string[TotalPages() - 1];
                if (page != 0) Array.Copy(Pages, 0, pages, 0, page);
                if (page != pages.Length) Array.Copy(Pages, page + 1, pages, page, pages.Length - page);
                return new BookAndQuill(pages);
            }

            public BookAndQuill InsertPage(int page, string text)
            {
                if (page is < 0 or >= 50) throw new ArgumentOutOfRangeException(nameof(page));
                ArgumentNullException.ThrowIfNull(text);
                if (Encoding.UTF8.GetByteCount(text) > 256) throw new ArgumentOutOfRangeException(nameof(text));
                if (page > TotalPages()) throw new ArgumentOutOfRangeException(nameof(page));
                var pages = new string[TotalPages() + 1];
                if (page != 0) Array.Copy(Pages, 0, pages, 0, page);
                pages[page] = text;
                if (page != TotalPages()) Array.Copy(Pages, page, pages, page + 1, TotalPages() - page);
                return new BookAndQuill(pages);
            }

            public BookAndQuill SetPage(int page, string text)
            {
                if (page is < 0 or >= 50) throw new ArgumentOutOfRangeException(nameof(page));
                ArgumentNullException.ThrowIfNull(text);
                if (Encoding.UTF8.GetByteCount(text) > 256) throw new ArgumentOutOfRangeException(nameof(text));
                var pages = new string[Math.Max(TotalPages(), page + 1)];
                Array.Fill(pages, string.Empty);
                if (TotalPages() != 0) Array.Copy(Pages, pages, TotalPages());
                pages[page] = text;
                return new BookAndQuill(pages);
            }

            public BookAndQuill SwapPages(int pageOne, int pageTwo)
            {
                if (pageOne < 0) throw new ArgumentOutOfRangeException(nameof(pageOne));
                if (pageTwo < 0) throw new ArgumentOutOfRangeException(nameof(pageTwo));
                if (Math.Max(pageOne, pageTwo) >= TotalPages()) throw new ArgumentOutOfRangeException();
                var pages = (string[])Pages.Clone();
                (pages[pageOne], pages[pageTwo]) = (pages[pageTwo], pages[pageOne]);
                return new BookAndQuill(pages);
            }
        }

`)
		return
	}
	output.WriteString(`        public readonly struct WrittenBook : World.Item
        {
            public WrittenBook(string title, string author, WrittenBookGeneration generation, params string[] pages)
            {
                ArgumentNullException.ThrowIfNull(title);
                ArgumentNullException.ThrowIfNull(author);
                ArgumentNullException.ThrowIfNull(pages);
                _title = title;
                _author = author;
                Generation = generation;
                _pages = pages;
            }

            private readonly string? _title;
            private readonly string? _author;
            private readonly string[]? _pages;
            public string Title => _title ?? string.Empty;
            public string Author => _author ?? string.Empty;
            public WrittenBookGeneration Generation { get; }
            public string[] Pages => _pages ?? [];
            public int TotalPages() => Pages.Length;

            public (string Page, bool Ok) Page(int page) => page >= 0 && page < TotalPages()
                ? (Pages[page], true)
                : (string.Empty, false);
        }

`)
}

func generateItemCapabilities(output *bytes.Buffer, spec itemSpec) {
	output.WriteString("    internal readonly record struct ItemDurability(int MaxDurability, bool Persistent, Item.Stack BrokenStack);\n\n")
	output.WriteString("    internal static class ItemCapabilities\n    {\n")
	output.WriteString("        internal static int MaxCount(World.Item? item) => item switch\n        {\n")
	if spec.Bucket.Present {
		output.WriteString("            Item.Bucket value => value.MaxCount(),\n")
	}
	for _, definition := range spec.Types {
		if definition.Bucket {
			continue
		}
		for _, state := range definition.States {
			if state.Capability.MaxCount != 64 {
				fmt.Fprintf(output, "            %s => %d,\n", csharpItemPattern(definition, state, spec), state.Capability.MaxCount)
			}
		}
	}
	output.WriteString("            _ => 64,\n        };\n\n")
	output.WriteString("        internal static bool TryDurability(World.Item? item, out ItemDurability durability)\n        {\n            switch (item)\n            {\n")
	for _, definition := range spec.Types {
		for _, state := range definition.States {
			capability := state.Capability
			if capability.MaxDurability < 0 {
				continue
			}
			broken := "default"
			if capability.BrokenIdentifier != "" && capability.BrokenCount > 0 {
				broken = fmt.Sprintf("new Item.Stack(ItemCodec.Decode(%s, %d), %d)", strconv.Quote(capability.BrokenIdentifier), capability.BrokenMetadata, capability.BrokenCount)
			}
			fmt.Fprintf(output, "                case %s:\n                    durability = new(%d, %s, %s); return true;\n",
				csharpItemPattern(definition, state, spec), capability.MaxDurability, strconv.FormatBool(capability.Persistent), broken)
		}
	}
	output.WriteString("                default:\n                    durability = default; return false;\n            }\n        }\n\n")
	output.WriteString("        internal static double AttackDamage(World.Item? item) => item switch\n        {\n")
	for _, definition := range spec.Types {
		for _, state := range definition.States {
			if state.Capability.AttackDamage != 1 {
				fmt.Fprintf(output, "            %s => %s,\n", csharpItemPattern(definition, state, spec), csharpDouble(state.Capability.AttackDamage))
			}
		}
	}
	output.WriteString("            _ => 1d,\n        };\n\n")
	output.WriteString("        internal static Item.FuelInfo FuelInfo(World.Item? item) => item switch\n        {\n")
	if spec.Bucket.Present {
		output.WriteString("            Item.Bucket value => value.FuelInfo(),\n")
	}
	for _, definition := range spec.Types {
		if definition.Bucket {
			continue
		}
		for _, state := range definition.States {
			capability := state.Capability
			if !capability.Fuel {
				continue
			}
			residue := "default"
			if capability.FuelIdentifier != "" && capability.FuelCount > 0 {
				residue = fmt.Sprintf("Item.NewStack(ItemCodec.Decode(%s, %d), %d)", strconv.Quote(capability.FuelIdentifier), capability.FuelMetadata, capability.FuelCount)
			}
			fmt.Fprintf(output, "            %s => new(TimeSpan.FromTicks(%d), %s),\n",
				csharpItemPattern(definition, state, spec), csharpDurationTicks(capability.FuelDuration), residue)
		}
	}
	output.WriteString("            _ => default,\n        };\n\n")
	output.WriteString("        internal static bool AllowsAnvilCost(World.Item? item) => item switch\n        {\n")
	for _, definition := range spec.Types {
		for _, state := range definition.States {
			if state.Capability.AllowsAnvilCost {
				fmt.Fprintf(output, "            %s => true,\n", csharpItemPattern(definition, state, spec))
			}
		}
	}
	output.WriteString("            _ => false,\n        };\n\n")
	output.WriteString("        internal static bool EnchantmentCompatibleWithItem(int enchantment, World.Item? item)\n        {\n")
	output.WriteString("            if (item is null || !ItemCodec.TryEncode(item, out var identifier, out var metadata)) return false;\n")
	output.WriteString("            return (enchantment, identifier, metadata) switch\n            {\n")
	for _, enchantment := range spec.Enchantments {
		for _, item := range enchantment.CompatibleItems {
			fmt.Fprintf(output, "                (%d, %s, %d) => true,\n", enchantment.ID, strconv.Quote(item.Identifier), item.Metadata)
		}
	}
	output.WriteString("                _ => false,\n            };\n        }\n")
	output.WriteString("    }\n")
}

func itemTypeFuel(definition itemTypeSpec) bool {
	if len(definition.States) == 0 {
		return false
	}
	fuel := definition.States[0].Capability.Fuel
	for _, state := range definition.States[1:] {
		if state.Capability.Fuel != fuel {
			panic("item type has inconsistent Fuel implementation: " + definition.Name)
		}
	}
	return fuel
}

func csharpDurationTicks(duration time.Duration) int64 {
	const tick = 100 * time.Nanosecond
	if duration%tick != 0 {
		panic("item duration does not fit C# TimeSpan ticks")
	}
	return int64(duration / tick)
}

func csharpItemPattern(definition itemTypeSpec, state itemStateSpec, spec itemSpec) string {
	if definition.Bucket {
		return csharpBucketPattern(state.Bucket)
	}
	if definition.Armour {
		return fmt.Sprintf("Item.%s value when value.Tier is Item.%s", definition.Name, spec.Armour.Tiers[state.ArmourTier].Name)
	}
	if definition.Name == "FireworkStar" {
		valueType := findItemValueType(spec.ValueTypes, "Colour")
		return fmt.Sprintf("Item.FireworkStar value when value.FireworkExplosion.Colour == %s", itemValueFactory(*valueType, state.Values[0]))
	}
	if len(definition.Fields) == 0 {
		return "Item." + definition.Name + " _"
	}
	field := definition.Fields[0]
	switch field.Kind {
	case itemFieldToolTier:
		return fmt.Sprintf("Item.%s value when value.%s == Item.%s", definition.Name, field.Name, spec.ToolTiers[state.ToolTier].Variable)
	case itemFieldValue:
		valueType := findItemValueType(spec.ValueTypes, field.ValueType)
		return fmt.Sprintf("Item.%s value when value.%s == %s", definition.Name, field.Name, itemValueFactory(*valueType, state.Values[0]))
	default:
		parts := make([]string, len(definition.Fields))
		for index, field := range definition.Fields {
			parts[index] = fmt.Sprintf("%s: %t", field.Name, state.Bools[index])
		}
		return fmt.Sprintf("Item.%s { %s }", definition.Name, strings.Join(parts, ", "))
	}
}

func generateItemValueType(output *bytes.Buffer, spec itemValueTypeSpec, indent string) {
	fmt.Fprintf(output, "%spublic readonly record struct %s\n%s{\n", indent, spec.Name, indent)
	fmt.Fprintf(output, "%s    private readonly int _value;\n", indent)
	fmt.Fprintf(output, "%s    internal %s(int value) => _value = value;\n", indent, spec.Name)
	fmt.Fprintf(output, "%s    internal int Id => _value;\n", indent)
	for _, method := range spec.Methods {
		fmt.Fprintf(output, "\n%s    public %s %s() => _value switch\n%s    {\n", indent, method.ReturnType, method.Name, indent)
		for index, result := range method.Results {
			fmt.Fprintf(output, "%s        %d => %s,\n", indent, index, result)
		}
		if method.Default != "" {
			fmt.Fprintf(output, "%s        _ => %s,\n%s    };\n", indent, method.Default, indent)
		} else {
			fmt.Fprintf(output, "%s        _ => throw new InvalidOperationException(\"Invalid %s value.\"),\n%s    };\n", indent, spec.Name, indent)
		}
	}
	fmt.Fprintf(output, "%s}\n\n", indent)
	for index, factory := range spec.Factories {
		fmt.Fprintf(output, "%spublic static %s %s() => new(%d);\n", indent, spec.Name, factory, index)
	}
	if spec.From {
		fmt.Fprintf(output, "\n%spublic static %s From(int id) => new(unchecked((byte)id));\n", indent, spec.Name)
	}
	if spec.Collection == "All" || spec.Collection == "StewTypes" {
		fmt.Fprintf(output, "\n%spublic static IReadOnlyList<%s> %s() => new %s[]\n%s{\n", indent, spec.Name, spec.Collection, spec.Name, indent)
		for _, factory := range spec.Factories {
			fmt.Fprintf(output, "%s    %s(),\n", indent, factory)
		}
		fmt.Fprintf(output, "%s};\n", indent)
	}
	output.WriteByte('\n')
}

func itemValueFactory(spec itemValueTypeSpec, index int) string {
	return spec.Container + "." + spec.Factories[index] + "()"
}

func csharpToolTier(tier dfitem.ToolTier) string {
	return strings.Join([]string{
		strconv.Itoa(tier.HarvestLevel),
		csharpDouble(tier.BaseMiningEfficiency),
		csharpDouble(tier.BaseAttackDamage),
		strconv.Itoa(tier.EnchantmentValue),
		strconv.Itoa(tier.Durability),
		strconv.Quote(tier.Name),
	}, ", ")
}

func csharpDouble(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, 64) + "d"
}

func generateGameModes(spec gameModeSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/world/game_mode.go AST and live registry. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\nusing System;\n\nnamespace Dragonfly;\n\n")
	output.WriteString("public sealed partial class World\n{\n")
	output.WriteString("    public interface GameMode\n    {\n")
	for _, method := range spec.Methods {
		fmt.Fprintf(&output, "        bool %s();\n", method)
	}
	output.WriteString("    }\n\n")
	for _, mode := range spec.Modes {
		fmt.Fprintf(&output, "    public static readonly GameMode %s = new BuiltinGameMode(%d, 0x%02xUL);\n",
			mode.Name, mode.ID, gameModeCapabilityMask(mode.Capabilities))
	}
	output.WriteString("\n    public static (GameMode GameMode, bool Ok) GameModeByID(int id) => id switch\n    {\n")
	for _, mode := range spec.Modes {
		fmt.Fprintf(&output, "        %d => (%s, true),\n", mode.ID, mode.Name)
	}
	output.WriteString("        _ => (GameModeSurvival, false),\n    };\n\n")
	output.WriteString(`    public static (int ID, bool Ok) GameModeID(GameMode mode)
    {
        if (mode is BuiltinGameMode builtin) return (builtin.ID, true);
        return (0, false);
    }

    internal static long GameModeDescriptor(GameMode mode)
    {
        ArgumentNullException.ThrowIfNull(mode);
        if (mode is BuiltinGameMode builtin)
            return unchecked((long)(BuiltinGameModeFlag | (uint)builtin.ID));
        ulong capabilities = 0;
`)
	for index, method := range spec.Methods {
		fmt.Fprintf(&output, "        if (mode.%s()) capabilities |= 1UL << %d;\n", method, index)
	}
	output.WriteString(`        return (long)capabilities;
    }

    internal static GameMode GameModeFromDescriptor(long descriptor)
    {
        var value = unchecked((ulong)descriptor);
        if ((value & BuiltinGameModeFlag) != 0)
        {
            var rawID = value & ~BuiltinGameModeFlag;
            if (rawID > int.MaxValue)
                throw new InvalidOperationException("invalid game mode descriptor");
            var (mode, ok) = GameModeByID((int)rawID);
            if (!ok) throw new InvalidOperationException("invalid game mode descriptor");
            return mode;
        }
        if ((value & ~CustomGameModeMask) != 0)
            throw new InvalidOperationException("invalid game mode descriptor");
        return new CapabilityGameMode(value);
    }

    private const ulong BuiltinGameModeFlag = 1UL << 63;
    private const ulong CustomGameModeMask = (1UL << 8) - 1;

    private class CapabilityGameMode(ulong capabilities) : GameMode
    {
`)
	for index, method := range spec.Methods {
		fmt.Fprintf(&output, "        public bool %s() => (capabilities & (1UL << %d)) != 0;\n", method, index)
	}
	output.WriteString(`    }

    private sealed class BuiltinGameMode(int id, ulong capabilities) : CapabilityGameMode(capabilities)
    {
        internal int ID { get; } = id;
    }
}
`)
	return output.Bytes()
}

func gameModeCapabilityMask(capabilities []bool) uint64 {
	var value uint64
	for index, enabled := range capabilities {
		if enabled {
			value |= 1 << index
		}
	}
	return value
}

func generatePlayerGameModes(methods []commandMethod) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\nusing System;\nusing Dragonfly.Native;\n\nnamespace Dragonfly;\n\n")
	output.WriteString("public sealed partial class Player\n{\n")
	for index, method := range methods {
		switch method.Name {
		case "SetGameMode":
			output.WriteString(`    public void SetGameMode(World.GameMode mode)
    {
        ArgumentNullException.ThrowIfNull(mode);
        PluginBridge.Host.SetPlayerState(
            _invocation,
            Id,
            Abi.PlayerStateGameMode,
            new PlayerStateValue { Integer = World.GameModeDescriptor(mode) });
    }
`)
		case "GameMode":
			output.WriteString("    public World.GameMode GameMode() => PluginBridge.Host.PlayerGameMode(_invocation, Id);\n")
		default:
			panic("unsupported player game mode method: " + method.Name)
		}
		if index != len(methods)-1 {
			output.WriteByte('\n')
		}
	}
	output.WriteString("}\n")
	return output.Bytes()
}

func generateParticles(spec particleSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly particle, sound/instrument, and image/color Go AST. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\n\n")
	output.WriteString("namespace Dragonfly\n{\n")
	output.WriteString("    public static partial class Color\n    {\n")
	fmt.Fprintf(&output, "        public readonly record struct RGBA(%s);\n", formatParameters(spec.RGBAFields))
	output.WriteString("    }\n\n")
	output.WriteString("    public static partial class Sound\n    {\n")
	output.WriteString("        public readonly struct Instrument\n        {\n")
	output.WriteString("            private readonly uint _id;\n")
	output.WriteString("            internal Instrument(uint id) => _id = id;\n")
	output.WriteString("            internal uint Id => _id;\n")
	output.WriteString("        }\n\n")
	for _, instrument := range spec.Instruments {
		fmt.Fprintf(&output, "        public static Instrument %s() => new(%du);\n", instrument.Name, instrument.ID)
	}
	output.WriteString("    }\n\n")
	output.WriteString("    public static partial class Particle\n    {\n")
	for _, particle := range spec.Types {
		if len(particle.Fields) == 0 {
			fmt.Fprintf(&output, "        public readonly record struct %s : World.Particle;\n", particle.Name)
			continue
		}
		fmt.Fprintf(&output, "        public readonly record struct %s(%s) : World.Particle;\n",
			particle.Name, formatParameters(particle.Fields))
	}
	output.WriteString("    }\n\n")
	output.WriteString("    internal readonly record struct EncodedParticle(\n")
	output.WriteString("        uint Kind, uint Data, int Pitch, Color.RGBA Colour, Cube.Pos Diff, World.Block? Block);\n\n")
	output.WriteString("    internal static class ParticleCodec\n    {\n")
	output.WriteString("        internal static bool TryEncode(World.Particle particle, out EncodedParticle encoded)\n        {\n")
	output.WriteString("            switch (particle)\n            {\n")
	for _, particle := range spec.Types {
		binding := "_"
		if len(particle.Fields) != 0 {
			binding = "value"
		}
		fmt.Fprintf(&output, "                case Particle.%s %s:\n", particle.Name, binding)
		fmt.Fprintf(&output, "                    encoded = new(%du, %s); return true;\n", particle.Kind, particleEncodedArguments(particle))
	}
	output.WriteString("                default:\n                    encoded = default; return false;\n")
	output.WriteString("            }\n        }\n    }\n}\n")
	return output.Bytes()
}

func particleEncodedArguments(particle particleType) string {
	values := map[string]string{
		"data":   "0u",
		"pitch":  "0",
		"colour": "default",
		"diff":   "default",
		"block":  "null",
	}
	for _, field := range particle.Fields {
		expression := "value." + field.Name
		switch field.Type {
		case "bool":
			values["data"] = expression + " ? 1u : 0u"
		case "int":
			values["pitch"] = expression
		case "Color.RGBA":
			values["colour"] = expression
		case "Cube.Face":
			values["data"] = "(uint)" + expression
		case "Cube.Pos":
			values["diff"] = expression
		case "Sound.Instrument":
			values["data"] = expression + ".Id"
		case "World.Block":
			values["block"] = expression
		default:
			panic("unsupported particle field type: " + field.Type)
		}
	}
	return strings.Join([]string{values["data"], values["pitch"], values["colour"], values["diff"], values["block"]}, ", ")
}

func csharpBytes(value []byte) string {
	if len(value) == 0 {
		return "Array.Empty<byte>()"
	}
	parts := make([]string, len(value))
	for index, current := range value {
		parts[index] = fmt.Sprintf("0x%02x", current)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func formatParameters(parameters []parameter) string {
	values := make([]string, len(parameters))
	for index, parameter := range parameters {
		values[index] = parameter.Type + " " + parameter.Name
	}
	return strings.Join(values, ", ")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
