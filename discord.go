package main

type DiscordTeamMember struct {
	// the user's membership state on the team
	Membership_state int `json:"membership_state"`
	// will always be ["*"]
	Permissions []string `json:"permissions"`
	// the id of the parent team of which they are a member
	TeamId string `json:"team_id"`
	// the avatar, discriminator, id, and username of the user
	User DiscordUser `json:"user"`
}

type DiscordTeam struct {
	// a hash of the image of the team's icon
	Icon string `json:"icon"`
	// the unique id of the team
	Id string `json:"id"`
	// the members of the team
	Members []DiscordTeamMember `json:"members"`
	// the user id of the current team owner
	OwnerUserId string `json:"owner_user_id"`
}

type DiscordUser struct {
	// the user's id identify
	Id string `json:"id"`
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

type DiscordResponse struct {
	// the id of the app
	Id string `json:"id"`
	// the name of the app
	Name string `json:"name"`
	// the icon hash of the app
	Icon string `json:"icon"`
	// the description of the app
	Description string `json:"description"`
	// an array of rpc origin urls, if rpc is enabled
	RpcOrigins []string `json:"rpc_origins"`
	// when false only app owner can join the app's bot to guilds
	BotPublic bool `json:"bot_public"`
	// when true the app's bot will only join upon completion of the full oauth2 code grant flow
	BotRequireCodeGrant bool `json:"bot_require_code_grant"`
	// partial user object containing info on the owner of the application
	Owner DiscordUser `json:"owner"`
	// if this application is a game sold on Discord, this field will be the summary field for the store page of its primary sku
	Summary string `json:"summary"`
	// the base64 encoded key for the GameSDK's GetTicket
	VerifyKey string `json:"verify_key"`
	// if the application belongs to a team, this will be a list of the members of that team
	Team DiscordTeam `json:"team"`
	// if this application is a game sold on Discord, this field will be the guild to which it has been linked
	GuildId string `json:"guild_id"`
	// if this application is a game sold on Discord, this field will be the id of the "Game SKU" that is created, if exists
	PrimarySkuId string `json:"primary_sku_id"`
	// if this application is a game sold on Discord, this field will be the URL slug that links to the store page
	Slug string `json:"slug"`
	// if this application is a game sold on Discord, this field will be the hash of the image on store embeds
	CoverImage string `json:"cover_image"`
	// the application's public flags
	Flags int `json:"flags"`
}
