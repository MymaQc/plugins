// Code generated from Dragonfly Go AST and live registries by csharp-gen. DO NOT EDIT.

package native

import "time"

type PlayerStateKind uint32

const (
	PlayerStateGameMode            PlayerStateKind = 0
	PlayerStateFood                PlayerStateKind = 3
	PlayerStateMaxHealth           PlayerStateKind = 4
	PlayerStateHealth              PlayerStateKind = 5
	PlayerStateExperienceLevel     PlayerStateKind = 6
	PlayerStateExperienceProgress  PlayerStateKind = 7
	PlayerStateScale               PlayerStateKind = 8
	PlayerStateInvisible           PlayerStateKind = 9
	PlayerStateImmobile            PlayerStateKind = 10
	PlayerStateSpeed               PlayerStateKind = 11
	PlayerStateFlightSpeed         PlayerStateKind = 12
	PlayerStateVerticalFlightSpeed PlayerStateKind = 13
	PlayerStateFallDistance        PlayerStateKind = 14
	PlayerStateAbsorption          PlayerStateKind = 15
	PlayerStateDead                PlayerStateKind = 16
	PlayerStateOnGround            PlayerStateKind = 17
	PlayerStateEyeHeight           PlayerStateKind = 18
	PlayerStateTorsoHeight         PlayerStateKind = 19
	PlayerStateBreathing           PlayerStateKind = 20
	PlayerStateSprinting           PlayerStateKind = 21
	PlayerStateSneaking            PlayerStateKind = 22
	PlayerStateSwimming            PlayerStateKind = 23
	PlayerStateCrawling            PlayerStateKind = 24
	PlayerStateGliding             PlayerStateKind = 25
	PlayerStateFlying              PlayerStateKind = 26
	PlayerStateOnFireDuration      PlayerStateKind = 27
	PlayerStateFireProof           PlayerStateKind = 28
	PlayerStateAirSupply           PlayerStateKind = 29
	PlayerStateMaxAirSupply        PlayerStateKind = 30
)

type PlayerStateValue struct {
	Number  float64
	Integer int64
}

type EffectType int32

const (
	EffectSpeed          EffectType = 1
	EffectSlowness       EffectType = 2
	EffectHaste          EffectType = 3
	EffectMiningFatigue  EffectType = 4
	EffectStrength       EffectType = 5
	EffectInstantHealth  EffectType = 6
	EffectInstantDamage  EffectType = 7
	EffectJumpBoost      EffectType = 8
	EffectNausea         EffectType = 9
	EffectRegeneration   EffectType = 10
	EffectResistance     EffectType = 11
	EffectFireResistance EffectType = 12
	EffectWaterBreathing EffectType = 13
	EffectInvisibility   EffectType = 14
	EffectBlindness      EffectType = 15
	EffectNightVision    EffectType = 16
	EffectHunger         EffectType = 17
	EffectWeakness       EffectType = 18
	EffectPoison         EffectType = 19
	EffectWither         EffectType = 20
	EffectHealthBoost    EffectType = 21
	EffectAbsorption     EffectType = 22
	EffectSaturation     EffectType = 23
	EffectLevitation     EffectType = 24
	EffectFatalPoison    EffectType = 25
	EffectConduitPower   EffectType = 26
	EffectSlowFalling    EffectType = 27
	EffectDarkness       EffectType = 30
)

type PlayerEffectOperation uint32

const (
	PlayerEffectAdd PlayerEffectOperation = iota
	PlayerEffectRemove
)

type PlayerEffect struct {
	Type            EffectType
	Level           int32
	Duration        time.Duration
	Potency         float64
	Ambient         bool
	ParticlesHidden bool
	Infinite        bool
	Tick            int64
}

type PlayerTextKind uint32

const (
	PlayerTextMessage      PlayerTextKind = 0
	PlayerTextTip          PlayerTextKind = 1
	PlayerTextPopup        PlayerTextKind = 2
	PlayerTextJukeboxPopup PlayerTextKind = 3
	PlayerTextNameTag      PlayerTextKind = 4
	PlayerTextDisconnect   PlayerTextKind = 5
)

type SoundKind uint32

const (
	SoundAnvilBreak                 SoundKind = 0
	SoundAnvilLand                  SoundKind = 1
	SoundAnvilUse                   SoundKind = 2
	SoundArrowHit                   SoundKind = 3
	SoundBarrelClose                SoundKind = 4
	SoundBarrelOpen                 SoundKind = 5
	SoundBlastFurnaceCrackle        SoundKind = 6
	SoundBowShoot                   SoundKind = 7
	SoundBurning                    SoundKind = 8
	SoundBurp                       SoundKind = 9
	SoundCampfireCrackle            SoundKind = 10
	SoundChestClose                 SoundKind = 11
	SoundChestOpen                  SoundKind = 12
	SoundClick                      SoundKind = 13
	SoundComposterEmpty             SoundKind = 14
	SoundComposterFill              SoundKind = 15
	SoundComposterFillLayer         SoundKind = 16
	SoundComposterReady             SoundKind = 17
	SoundCopperScraped              SoundKind = 18
	SoundCrossbowShoot              SoundKind = 19
	SoundDecoratedPotInsertFailed   SoundKind = 20
	SoundDeny                       SoundKind = 21
	SoundDoorCrash                  SoundKind = 22
	SoundDrowning                   SoundKind = 23
	SoundEnderChestClose            SoundKind = 24
	SoundEnderChestOpen             SoundKind = 25
	SoundExperience                 SoundKind = 26
	SoundExplosion                  SoundKind = 27
	SoundFireCharge                 SoundKind = 28
	SoundFireExtinguish             SoundKind = 29
	SoundFireworkBlast              SoundKind = 30
	SoundFireworkHugeBlast          SoundKind = 31
	SoundFireworkLaunch             SoundKind = 32
	SoundFireworkTwinkle            SoundKind = 33
	SoundFizz                       SoundKind = 34
	SoundFurnaceCrackle             SoundKind = 35
	SoundGhastShoot                 SoundKind = 36
	SoundGhastWarning               SoundKind = 37
	SoundGlassBreak                 SoundKind = 38
	SoundIgnite                     SoundKind = 39
	SoundItemAdd                    SoundKind = 40
	SoundItemBreak                  SoundKind = 41
	SoundItemFrameRemove            SoundKind = 42
	SoundItemFrameRotate            SoundKind = 43
	SoundItemThrow                  SoundKind = 44
	SoundLecternBookPlace           SoundKind = 45
	SoundLevelUp                    SoundKind = 46
	SoundLightningExplode           SoundKind = 47
	SoundLightningThunder           SoundKind = 48
	SoundMusicDiscEnd               SoundKind = 49
	SoundPop                        SoundKind = 50
	SoundPotionBrewed               SoundKind = 51
	SoundPowerOff                   SoundKind = 52
	SoundPowerOn                    SoundKind = 53
	SoundSignWaxed                  SoundKind = 54
	SoundSmokerCrackle              SoundKind = 55
	SoundStopUsingSpyglass          SoundKind = 56
	SoundTnt                        SoundKind = 57
	SoundTeleport                   SoundKind = 58
	SoundThunder                    SoundKind = 59
	SoundTotem                      SoundKind = 60
	SoundUseSpyglass                SoundKind = 61
	SoundWaxRemoved                 SoundKind = 62
	SoundWaxedSignFailedInteraction SoundKind = 63
	SoundShulkerBoxOpen             SoundKind = 64
	SoundShulkerBoxClose            SoundKind = 65
	SoundEnderEyePlaced             SoundKind = 66
	SoundEndPortalCreated           SoundKind = 67
	SoundAttack                     SoundKind = 68
	SoundFall                       SoundKind = 69
	SoundBlockPlace                 SoundKind = 70
	SoundBlockBreaking              SoundKind = 71
	SoundDoorOpen                   SoundKind = 72
	SoundDoorClose                  SoundKind = 73
	SoundTrapdoorOpen               SoundKind = 74
	SoundTrapdoorClose              SoundKind = 75
	SoundFenceGateOpen              SoundKind = 76
	SoundFenceGateClose             SoundKind = 77
	SoundNote                       SoundKind = 78
	SoundMusicDiscPlay              SoundKind = 79
	SoundDecoratedPotInserted       SoundKind = 80
	SoundItemUseOn                  SoundKind = 81
	SoundEquipItem                  SoundKind = 82
	SoundBucketFill                 SoundKind = 83
	SoundBucketEmpty                SoundKind = 84
	SoundCrossbowLoad               SoundKind = 85
	SoundGoatHorn                   SoundKind = 86
)
