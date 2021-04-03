package discord

type TeamMember struct {
	// the user's membership state on the team
	MembershipState int `json:"membership_state"`
	// will always be ["*"]
	Permissions []string `json:"permissions"`
	// the id of the parent team of which they are a member
	TeamID string `json:"team_id"`
	// the avatar, discriminator, id, and username of the user
	User User `json:"user"`
}

type Team struct {
	// a hash of the image of the team's icon
	Icon string `json:"icon"`
	// the unique id of the team
	ID string `json:"id"`
	// the members of the team
	Members []TeamMember `json:"members"`
	// the user id of the current team owner
	OwnerUserID string `json:"owner_user_id"`
}

type User struct {
	// the user's id identify
	ID string `json:"id"`
	// the user's username, not unique across the platform identify
	Username string `json:"username"`
	// the user's 4-digit discord-tag identify
	Discriminator string `json:"discriminator"`
	// the user's avatar hash identify
	Avatar string `json:"avatar"`
	// whether the user belongs to an OAuth2 application identify
	Bot bool `json:"bot"`
	// whether the user is an Official Discord System user (part of the urgent message system) identify
	System bool `json:"system"`
	// whether the user has two factor enabled on their account identify
	MfaEnabled bool `json:"mfa_enabled"`
	// the user's chosen language option identify
	Locale string `json:"locale"`
	// whether the email on this account has been verified email
	Verified bool `json:"verified"`
	// the user's email email
	Email string `json:"email"`
	// the flags on a user's account identify
	Flags int `json:"flags"`
	// the type of Nitro subscription on a user's account identify
	PremiumType int `json:"premium_type"`
	// the public flags on a user's account identify
	PublicFlags int `json:"public_flags"`
}

type Application struct {
	// the id of the app
	ID string `json:"id"`
	// the name of the app
	Name string `json:"name"`
	// the icon hash of the app
	Icon string `json:"icon"`
	// the description of the app
	Description string `json:"description"`
	// an array of rpc origin urls, if rpc is enabled
	RPCOrigins []string `json:"rpc_origins"`
	// when false only app owner can join the app's bot to guilds
	BotPublic bool `json:"bot_public"`
	// when true the app's bot will only join upon completion of the full oauth2 code grant flow
	BotRequireCodeGrant bool `json:"bot_require_code_grant"`
	// partial user object containing info on the owner of the application
	Owner   User   `json:"owner"`
	Summary string `json:"summary"`
	// if this application is a game sold on Discord, this field will be the summary field for the store page of its primary sku
	// the base64 encoded key for the GameSDK's GetTicket
	VerifyKey string `json:"verify_key"`
	// if the application belongs to a team, this will be a list of the members of that team
	Team Team `json:"team"`
	// if this application is a game sold on Discord, this field will be the guild to which it has been linked
	GuildID string `json:"guild_id"`
	// if this application is a game sold on Discord, this field will be the id of the "Game SKU" that is created, if exists
	PrimarySkuID string `json:"primary_sku_id"`
	// if this application is a game sold on Discord, this field will be the URL slug that links to the store page
	Slug string `json:"slug"`
	// if this application is a game sold on Discord, this field will be the hash of the image on store embeds
	CoverImage string `json:"cover_image"`
	// the application's public flags
	Flags int `json:"flags"`
}

type Guild struct {
	// * These fields are only sent within the GUILDCREATE event
	// ** These fields are only sent when using the GET Current User Guilds endpoint and are relative to the requested user
	// guild id
	ID string `json:"id"`
	// guild name (2-100 characters, excluding trailing and leading whitespace)
	Name string `json:"name"`
	// icon hash
	Icon string `json:"icon"`
	// icon hash, returned when in the template object
	IconHash string `json:"icon_hash"`
	// splash hash
	Splash string `json:"splash"`
	// discovery splash hash; only present for guilds with the "DISCOVERABLE" feature
	DiscoverySplash string `json:"discovery_splash"`
	// true if the user is the owner of the guild **
	Owner bool `json:"owner"`
	// id of owner
	OwnerID string `json:"owner_id"`
	// total permissions for the user in the guild (excludes overrides) **
	Permissions string `json:"permissions"`
	// voice region id for the guild
	Region string `json:"region"`
	// id of afk channel
	AfkChannelID string `json:"afk_channel_id"`
	// afk timeout in seconds
	AfkTimeout int `json:"afk_timeout"`
	// true if the server widget is enabled
	WidgetEnabled bool `json:"widget_enabled"`
	// the channel id that the widget will generate an invite to, or null if set to no invite
	WidgetChannelID string `json:"widget_channel_id"`
	// verification level required for the guild
	VerificationLevel int `json:"verification_level"`
	// default message notifications level
	DefaultMessageNotifications int `json:"default_message_notifications"`
	// explicit content filter level
	ExplicitContentFilter int `json:"explicit_content_filter"`
	// roles in the guild
	Roles []Role
	// custom guild emojis
	Emojis []Emoji
	// enabled guild features
	Features []string
	// required MFA level for the guild
	MfaLevel int `json:"mfa_level"`
	// application id of the guild creator if it is bot-created
	ApplicationID string `json:"application_id"`
	// the id of the channel where guild notices such as welcome messages and boost events are posted
	SystemChannelID string `json:"system_channel_id"`
	// system channel flags
	SystemChannelFlags int `json:"system_channel_flags"`
	// the id of the channel where Community guilds can display rules and/or guidelines
	RulesChannelID string `json:"rules_channel_id"`
	// when this guild was joined at *
	//  ISO8601 timestamp
	JoinedAt string `json:"joined_at"`
	// true if this is considered a large guild *
	Large bool `json:"large"`
	// true if this guild is unavailable due to an outage *
	Unavailable bool `json:"unavailable"`
	// total number of members in this guild *
	MemberCount int `json:"member_count"`
	// states of members currently in voice channels; lacks the guildID key *
	VoiceStates []VoiceState
	// users in the guild *
	Members []GuildMember
	// channels in the guild *
	Channels []Channel
	// presences of the members in the guild, will only include non-offline members if the size is
	// greater than large threshold *
	Presences []PrescenceUpdateEvent
	// the maximum number of presences for the guild (the default value, currently 25000, is in effect when null is returned)
	MaxPresences int `json:"max_presences"`
	// the maximum number of members for the guild
	MaxMembers int `json:"max_members"`
	// the vanity url code for the guild
	VanityURLCode string `json:"vanity_url_code"`
	// the description for the guild, if the guild is discoverable
	Description string `json:"description"`
	// banner hash
	Banner string `json:"banner"`
	// premium tier (Server Boost level)
	PremiumTier int `json:"premium_tier"`
	// the number of boosts this guild currently has
	PremiumSubscriptionCount int `json:"premium_subscription_count"`
	// the preferred locale of a Community guild; used in server discovery and notices from Discord; defaults to "en-US"
	PreferredLocale string `json:"preferred_locale"`
	// the id of the channel where admins and moderators of Community guilds receive notices from Discord
	PublicUpdatesChannelID string `json:"public_updates_channel_id"`
	// the maximum amount of users in a video channel
	MaxVideoChannelUsers int `json:"max_video_channel_users"`
	// approximate number of members in this guild, returned from the GET /guilds/<id> endpoint when withCounts is true
	ApproximateMemberCount int `json:"approximate_member_count"`
	// approximate number of non-offline members in this guild, returned from the GET /guilds/<id> endpoint when withCounts is true
	ApproximatePresenceCount int `json:"approximate_presence_count"`
	// the welcome screen of a Community guild, shown to new members, returned when in the invite object
	WelcomeScreen WelcomeScreen `json:"welcomeScreen"`
}

type WelcomeScreen struct {
	// the server description shown in the welcome screen
	Description string `json:"description"`
	// the channels shown in the welcome screen, up to 5
	WelcomeChannels []WelcomeChannel
}

type WelcomeChannel struct {
	// the channel's id
	ChannelID string `json:"channel_id"`
	// the description shown for the channel
	Description string `json:"description"`
	// the emoji id, if the emoji is custom
	EmojiID string `json:"emoji_id"`
	// the emoji name if custom, the unicode character if standard, or null if no emoji is set
	EmojiName string `json:"emoji_name"`
}

type GuildMember struct {
	// the user this guild member represents
	User User `json:"user"`
	// this users guild nickname
	Nick string `json:"nick"`
	// array of role object ids
	Roles []string
	// when the user joined the guild
	JoinedAt string `json:"joined_at"` //  ISO8601 timestamp
	// when the user started boosting the guild
	PremiumSince string `json:"premium_since"` //  ISO8601 timestamp
	// whether the user is deafened in voice channels
	Deaf bool `json:"deaf"`
	// whether the user is muted in voice channels
	Mute bool `json:"mute"`
	// whether the user has not yet passed the guild's Membership Screening requirements
	Pending bool `json:"pending"`
	// total permissions of the member in the channel, including overrides, returned when in the interaction object
	Permissions string `json:"permissions"`
}

type Emoji struct {
	// emoji id
	ID string
	// emoji name (can be null only in reaction emoji objects)
	Name string
	// roles this emoji is whitelisted to
	Roles []string
	// user that created this emoji
	User User
	// whether this emoji must be wrapped in colons
	RequireColons bool
	// whether this emoji is managed
	Managed bool
	// whether this emoji is animated
	Animated bool
	// whether this emoji can be used, may be false due to loss of Server Boosts
	Available bool
}

type Role struct {
	// role id
	ID string `json:"id"`
	// role name
	Name string `json:"name"`
	// integer representation of hexadecimal color code
	Color int `json:"color"`
	// if this role is pinned in the user listing
	Hoist bool `json:"hoist"`
	// position of this role
	Position int `json:"position"`
	// permission bit set
	Permissions string `json:"permissions"`
	// whether this role is managed by an integration
	Managed bool `json:"managed"`
	// whether this role is mentionable
	Mentionable bool `json:"mentionable"`
	// the tags this role has
	Tags RoleTags `json:"tags"`
}

type RoleTags struct {
	// the id of the bot this role belongs to
	BotID string `json:"bot_id"`
	// the id of the integration this role belongs to
	IntegrationID string `json:"integration_id"`
	// whether this is the guild's premium subscriber role
	PremiumSubscriber bool `json:"premium_subscriber"`
}

type VoiceState struct {
	// the guild id this voice state is for
	GuildID string `json:"guild_id"`
	// the channel id this user is connected to
	ChannelID string `json:"channel_id"`
	// the user id this voice state is for
	UserID string `json:"user_id"`
	// the guild member this voice state is for
	Member GuildMember `json:"member"`
	// the session id for this voice state
	SessionID string `json:"session_id"`
	// whether this user is deafened by the server
	Deaf bool `json:"deaf"`
	// whether this user is muted by the server
	Mute bool `json:"mute"`
	// whether this user is locally deafened
	SelfDeaf bool `json:"self_deaf"`
	// whether this user is locally muted
	SelfMute bool `json:"self_mute"`
	// whether this user is streaming using "Go Live"
	SelfStream bool `json:"self_stream"`
	// whether this user's camera is enabled
	SelfVideo bool `json:"self_video"`
	// whether this user is muted by the current user
	Suppress bool `json:"suppress"`
}

type Channel struct {
	// the id of this channel
	ID string `json:"id"`
	// the type of channel
	Type int `json:"type"`
	// the id of the guild
	GuildID string `json:"guild_id"`
	// sorting position of the channel
	Position int `json:"position"`
	// explicit permission overwrites for members and roles
	PermissionOverwrites []Overwrite
	// the name of the channel (2-100 characters)
	Name string `json:"name"`
	// the channel topic (0-1024 characters)
	Topic string `json:"topic"`
	// whether the channel is nsfw
	Nsfw bool `json:"nsfw"`
	// the id of the last message sent in this channel (may not point to an existing or valid message)
	LastMessageID string `json:"last_message_id"`
	// the bitrate (in bits) of the voice channel
	Bitrate int `json:"bitrate"`
	// the user limit of the voice channel
	UserLimit int `json:"user_limit"`
	// amount of seconds a user has to wait before sending another message (0-21600); bots, as well as users with the permission manageMessages or manage_channel, are unaffected
	RateLimitPerUser int `json:"rate_limit_per_user"`
	// the recipients of the DM
	Recipients []User `json:"recipients"`
	// icon hash
	Icon string `json:"icon"`
	// id of the DM creator
	OwnerID string `json:"owner_id"`
	// application id of the group DM creator if it is bot-created
	ApplicationID string `json:"application_id"`
	// id of the parent category for a channel (each parent category can contain up to 50 channels)
	ParentID string `json:"parent_id"`
	// when the last pinned message was pinned. This may be null in events such as GUILDCREATE when a message is not pinned.
	LastPinTimestamp string `json:"last_pin_timestamp"`
}

type Overwrite struct {
	// role or user id
	ID string `json:"id"`
	// either 0 (role) or 1 (member)
	Type int `json:"type"`
	// permission bit set
	Allow string `json:"allow"`
	// permission bit set
	Deny string `json:"deny"`
}

type PrescenceUpdateEvent struct {
	// the user presence is being updated for
	User User `json:"user"`
	// id of the guild
	GuildID string `json:"guild_id"`
	// either "idle", "dnd", "online", or "offline"
	Status string `json:"status"`
	// user's current activities
	Activities []Activity `json:"activities"`
	// user's platform-dependent status
	ClientStatus ClientStatus `json:"client_status"`
}

type ClientStatus struct {
	// the user's status set for an active desktop (Windows, Linux, Mac) application session
	Desktop string `json:"desktop"`
	// the user's status set for an active mobile (iOS, Android) application session
	Mobile string `json:"mobile"`
	// the user's status set for an active web (browser, bot account) application session
	Web string `json:"web"`
}

type Activity struct {
	// the activity's name
	Name string `json:"name"`
	// activity type
	Type int `json:"type"`
	// stream url, is validated when type is 1
	URL string `json:"url"`
	// unix timestamp of when the activity was added to the user's session
	CreatedAt int `json:"created_at"`
	// unix timestamps for start and/or end of the game
	Timestamps ActivityTimestamps `json:"timestamps"`
	// application id for the game
	ApplicationID string `json:"application_id"`
	// what the player is currently doing
	Details string `json:"details"`
	// the user's current party status
	State string `json:"state"`
	// the emoji used for a custom status
	Emoji Emoji `json:"emoji"`
	// information for the current party of the player
	Party ActivityParty `json:"party"`
	// images for the presence and their hover texts
	Assets ActivityAssets `json:"assets"`
	// secrets for Rich Presence joining and spectating
	Secrets ActivitySecrets `json:"secrets"`
	// whether or not the activity is an instanced game session
	Instance bool `json:"instance"`
	// activity flags ORd together, describes what the payload includes
	Flags int `json:"flags"`
}

type ActivityTimestamps struct {
	// unix time (in milliseconds) of when the activity started
	Start int `json:"start"`
	// unix time (in milliseconds) of when the activity ends
	End int `json:"end"`
}

type ActivityParty struct {
	// the id of the party
	ID string `json:"id"`
	// used to show the party's current and maximum size
	Size []int `json:"size"` //  array of two ints (currentSize, max_size)
}

type ActivityAssets struct {
	// the id for a large asset of the activity, usually a string
	LargeImage string `json:"large_image"`
	// text displayed when hovering over the large image of the activity
	LargeText string `json:"large_text"`
	// the id for a small asset of the activity, usually a string
	SmallImage string `json:"small_image"`
	// text displayed when hovering over the small image of the activity
	SmallText string `json:"small_text"`
}

type ActivitySecrets struct {
	// the secret for joining a party
	Join string `json:"join"`
	// the secret for spectating a game
	Spectate string `json:"spectate"`
	// the secret for a specific instanced match
	Match string `json:"match"`
}
