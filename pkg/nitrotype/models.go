package nitrotype

import (
	"encoding/json"
	"fmt"
)

type NTPlayerLegacy map[string]interface{}

// UserProfile contains data from the NT Racer profile.
type NTPlayer struct {
	UserID             int            `json:"userID"`
	Username           string         `json:"username"`
	Membership         string         `json:"membership"`
	DisplayName        string         `json:"displayName"`
	Title              string         `json:"title"`
	Experience         int            `json:"experience"`
	Level              int            `json:"level"`
	TeamID             *int           `json:"teamID"`
	LookingForTeam     int            `json:"lookingForTeam"`
	CarID              int            `json:"carID"`
	CarHueAngle        int            `json:"carHueAngle"`
	TotalCars          int            `json:"totalCars"`
	Nitros             int            `json:"nitros"`
	NitrosUsed         int            `json:"nitrosUsed"`
	RacesPlayed        int            `json:"racesPlaywed"`
	LongestSession     int            `json:"longestSession"`
	AvgSpeed           int            `json:"avgSpeed"`
	HighestSpeed       int            `json:"highestSpeed"`
	AllowFriendRequest int            `json:"allowFriendRequests"`
	ProfileViews       int            `json:"profileViews"`
	CreatedStamp       int            `json:"createdStamp"`
	Tag                *string        `json:"tag"`
	TagColor           *string        `json:"tagColor"`
	Garage             []string       `json:"garage"`
	Cars               []NTPlayerCar  `json:"cars"`
	Loot               []NTPlayerLoot `json:"loot"`
}

// TODO: Enum types for loots.
// TODO: Work out car type.

// Loot contains loot information (usually used on the UserProfile).
type NTPlayerLoot struct {
	LootID       int                `json:"lootID"`
	Type         string             `json:"type"`
	Name         string             `json:"name"`
	AssetKey     string             `json:"assetKey"`
	Options      NTPlayerLootOption `json:"options"`
	Equipped     int                `json:"equipped"`
	CreatedStamp int                `json:"createdStamp"`
}

// LootOption contains options about a loot.
type NTPlayerLootOption struct {
	Src    string `json:"src"`
	Type   string `json:"type"`
	Rarity string `json:"rarity"`
}

// Car contains car info and it's paint job.
type NTPlayerCar struct {
	CarID        int    `json:"carID"`
	Status       string `json:"status"`
	CarHueAngle  int    `json:"carHueAngle"`
	CreatedStamp int    `json:"createdStamp"`
}

func (c *NTPlayerCar) UnmarshalJSON(bs []byte) error {
	data := []interface{}{}
	err := json.Unmarshal(bs, &data)
	if err != nil {
		return err
	}
	carID, ok := data[0].(float64)
	if !ok {
		return fmt.Errorf("failed to get car id")
	}
	status, ok := data[1].(string)
	if !ok {
		return fmt.Errorf("failed to get status")
	}
	carHueAngle, ok := data[2].(float64)
	if !ok {
		return fmt.Errorf("failed to get car hue angle")
	}
	createdStamp, ok := data[3].(float64)
	if !ok {
		return fmt.Errorf("failed to get created stamp")
	}
	c.CarID = int(carID)
	c.Status = status
	c.CarHueAngle = int(carHueAngle)
	c.CreatedStamp = int(createdStamp)
	return nil
}

type NTGlobalsLegacy map[string]interface{}

type NTGlobals struct {
	ActionSeasons []ActiveSeason `json:"ACTIVE_SEASONS"`
	Achievements  struct {
		List  []AchievementListItem          `json:"LIST"`
		Group []AchievementGroupItem         `json:"GROUP"`
		Text  map[string]AchievementTextItem `json:"TEXT"`
	} `json:"ACHIEVEMENTS"`
	Cars                   []Car                             `json:"CARS"`
	Products               map[string]Product                `json:"PRODUCTS"`
	GlobalAlert            bool                              `json:"GLOBAL_ALERT"`
	Loot                   []Loot                            `json:"LOOT"`
	Shop                   []Shop                            `json:"SHOP"`
	Dealership             []Dealership                      `json:"DEALERSHIP"`
	Challenges             []Challenge                       `json:"CHALLENGES"`
	StartingCars           []int                             `json:"STARTING_CARS"`
	FriendLimits           FriendLimits                      `json:"FRIEND_LIMITS"`
	PageLabels             map[string]string                 `json:"PAGE_LABELS"`
	OneWayFriendIDs        []int                             `json:"ONE_WAY_FRIEND_IDS"`
	TeamInfo               TeamInfo                          `json:"TEAM_INFO"`
	SeasonLevels           SeasonLevels                      `json:"SEASON_LEVELS"`
	TeachersURL            string                            `json:"TEACHERS_URL"`
	Sites                  map[string]string                 `json:"SITES"`
	CarURL                 string                            `json:"CAR_URL"`
	CarPaintedURL          string                            `json:"CAR_PAINTED_URL"`
	CashSending            CashSending                       `json:"CASH_SENDING"`
	ScoreboardRankMimimums map[string]ScoreboardRankMimimums `json:"SCOREBOARD_RANK_MINIMUMS"`
	LootConfig             map[string]LootConfig             `json:"LOOT_CONFIG"`
	ChallengeTypes         map[string][]string               `json:"CHALLENGE_TYPES"`
	TopPlayers             []RankItem
	TopTeams               []RankItem
}

type ActiveSeason struct {
	SeasonID             int    `json:"sessionID"`
	Name                 string `json:"name"`
	StartStamp           int64  `json:"startStamp"`
	EndStamp             int64  `json:"endStamp"`
	ClassName            string `json:"className"`
	AchievementGroupID   int    `json:"achievementGroupID"`
	AchievementGroupName string `json:"achievementGroupName"`
	TotalRewards         int    `json:"totalRewards"`
}

type AchievementListItem struct {
	AchievementID   int                       `json:"achievementID"`
	GID             int                       `json:"gid"`
	RuleGroup       string                    `json:"ruleGroup"`
	Name            string                    `json:"name"`
	Points          int                       `json:"points"`
	Rules           []AchievementListItemRule `json:"rules"`
	Reward          AchievementListItemReward `json:"reward"`
	RewardDesc      string                    `json:"rewardDesc"`
	Hidden          int                       `json:"hidden"`
	Active          int                       `json:"active"`
	NewCarThumbs    *[]string                 `json:"newCarThumbs"`
	SeasonID        *int                      `json:"seasonID"`
	StartStamp      *int64                    `json:"startStamp"`
	EndStamp        *int64                    `json:"endStamp"`
	SeasonName      *string                   `json:"seasonName"`
	SeasonClassName *string                   `json:"seasonClassName"`
}

type AchievementListItemRule struct {
	Field      string      `json:"field"`
	Comparison string      `json:"comparison"`
	Value      interface{} `json:"value"`
}

type AchievementListItemReward struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type AchievementGroupItem struct {
	AchievementGroupID int     `json:"achievementGroupID"`
	Site               string  `json:"site"`
	Name               string  `json:"name"`
	Type               string  `json:"type"`
	SeasonID           int     `json:"seasonID"`
	Img                *string `json:"img"`
	DisplayOrder       int     `json:"displayOrder"`
	ID                 int     `json:"id"`
	Order              int     `json:"order"`
	SeasonClassName    *string `json:"seasonClassName"`
	StartStamp         *int64  `json:"startStamp"`
	EndStamp           *int64  `json:"endStamp"`
	SeasonName         *string `json:"seasonName"`
}

type AchievementTextItem struct {
	Text       string  `json:"text"`
	Prefix     *string `json:"prefix,omitempty"`
	Format     *string `json:"format,omitempty"`
	Deprecated *bool   `json:"deprecated,omitempty"`
}

type Car struct {
	ID              int        `json:"id"`
	AssetKey        *string    `json:"assetKey,omitempty"`
	CarID           int        `json:"carID"`
	Name            string     `json:"name"`
	LongDescription string     `json:"longDescription"`
	Options         CarOptions `json:"options"`
	EnterSound      string     `json:"enterSound"`
	Price           int64      `json:"price"`
	LastModified    int64      `json:"lastModified"`
}

type CarOptions struct {
	Rarity     string `json:"rarity"`
	LargeSrc   string `json:"largeSrc"`
	SmallSrc   string `json:"smallSrc"`
	IsAnimated int    `json:"isAnimated,omitempty"`
}

type Product struct {
	ProductID   int     `json:"productID"`
	SKU         string  `json:"SKU"`
	AssetKey    *string `json:"assetKey"`
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	Featured    int     `json:"featured"`
	Description string  `json:"description"`
	CashReward  int64   `json:"cashReward"`
	Price       string  `json:"price"`
	SalePrice   string  `json:"salePrice"`
	SaleEnds    int64   `json:"saleEnds"`
	Active      int     `json:"active"`
}

type Loot struct {
	LootID          int         `json:"lootID"`
	Type            string      `json:"type"`
	Name            string      `json:"name"`
	Options         LootOptions `json:"options"`
	LongDescription *string     `json:"longDescription,omitempty"`
	Price           *int64      `json:"price,omitempty"`
	LastModified    int64       `json:"lastModified"`
}

type LootOptions struct {
	Src          *string `json:"src,omitempty"`
	Type         *string `json:"type,omitempty"`
	Rarity       string  `json:"rarity"`
	ResourceType *string `json:"resourceType,omitempty"`
}

type Shop struct {
	Category      string     `json:"category"`
	StartStamp    int64      `json:"startStamp"`
	Expiration    int64      `json:"expiration"`
	Items         []ShopItem `json:"items"`
	ShopReleaseID int        `json:"shopReleaseID"`
}

type ShopItem struct {
	Type             string  `json:"type"`
	ID               int     `json:"id"`
	Price            *int64  `json:"price"`
	ShortDescription *string `json:"shortDescription"`
	LongDescription  *string `json:"longDescription"`
	SlrID            int     `json:"slrID"`
}

type Dealership struct {
	DealershipID int              `json:"dealershipID"`
	AssetKey     string           `json:"assetKey"`
	Name         string           `json:"name"`
	Expiration   *string          `json:"expiration"`
	Items        []DealershipItem `json:"items"`
}

type DealershipItem struct {
	Type             string  `json:"type"`
	ID               int     `json:"id"`
	Price            *int64  `json:"price"`
	ShortDescription *string `json:"shortDescription"`
	LongDescription  *string `json:"longDescription"`
	DlID             int     `json:"dlID"`
}

type Challenge struct {
	ChallengeID int    `json:"challengeID"`
	Duration    string `json:"duration"`
	Type        string `json:"type"`
	Reward      int    `json:"reward"`
	Goal        int    `json:"goal"`
	Expiration  int64  `json:"expiration"`
}

type FriendLimits struct {
	Basic int `json:"basic"`
	Gold  int `json:"gold"`
}

type TeamInfo struct {
	Price              int64 `json:"price"`
	MinRaces           int   `json:"minRaces"`
	MaxMembers         int   `json:"maxMembers"`
	MaxOfficers        int   `json:"maxOfficers"`
	MOTDUpdateInterval int   `json:"motdUpdateInterval"`
	AutoRemoveOptions  []int `json:"autoRemoveOptions"`
}

type SeasonLevels struct {
	StartingLevels                int   `json:"startingLevels"`
	ExperiencePerStartingLevel    int64 `json:"experiencePerStartingLevel"`
	ExperiencePerAchievementLevel int64 `json:"experiencePerAchievementLevel"`
	ExperiencePerExtraLevels      int64 `json:"experiencePerExtraLevels"`
	ExtraLevelReward              int64 `json:"extraLevelReward"`
}

type CashSending struct {
	MinLevel        int     `json:"minLevel"`
	Minimum         int64   `json:"minimum"`
	Maximum         int64   `json:"maximum"`
	MaxPerWeek      int64   `json:"maxPerWeek"`
	MaxPerWeekTeams int64   `json:"maxPerWeekTeams"`
	FeePercent      float64 `json:"feePercent"`
	MinAccountAge   int     `json:"minAccountAge"`
}

type ScoreboardRankMimimums struct {
	Season  int `json:"season"`
	Monthly int `json:"monthly"`
	Weekly  int `json:"weekly"`
	Daily   int `json:"daily"`
}

type LootConfig struct {
	Defaults    []int  `json:"defaults"`
	MaxEquipped int    `json:"maxEquipped"`
	Name        string `json:"name"`
}

type RankItem struct {
	ID       int `json:"id"`
	Position int `json:"position"`
}
