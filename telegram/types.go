package telegram

import "encoding/json"

type Bot struct {
	token       string
	apiEndpoint string
}

type Response struct {
	Ok                 bool                `json:"ok"`
	Description        *string             `json:"description"`
	Result             *json.RawMessage    `json:"result"`
	ErrorCode          int                 `json:"error_code"`
	ResponseParameters *ResponseParameters `json:"response_parameters"`
}

type ResponseParameters struct {
	MigrateToChatId *int `json:"migrate_to_chat_id"`
	RetryAfter      *int `json:"retry_after"`
}

type Update struct {
	UpdateId          int            `json:"update_id"`
	Message           *Message       `json:"message"`
	EditedMessage     *Message       `json:"edited_message"`
	ChannelPost       *Message       `json:"channel_post"`
	EditedChannelPost *Message       `json:"edited_channel_post"`
	CallbackQuery     *CallbackQuery `json:"callback_query"`
}

type User struct {
	Id                      int    `json:"id"`
	IsBot                   bool   `json:"is_bot"`
	FirstName               string `json:"first_name"`
	LastName                string `json:"last_name,omitempty"`
	Username                string `json:"username,omitempty"`
	LanguageCode            string `json:"language_code,omitempty"`
	CanJoinGroups           *bool  `json:"can_join_groups"`
	CanReadAllGroupMessages *bool  `json:"can_read_all_group_messages"`
	SupportsInlineQueries   *bool  `json:"supports_inline_queries"`
}

type Message struct {
	MessageId  int    `json:"message_id"`
	From       *User  `json:"from"`
	SenderChat *Chat  `json:"sender_chat"`
	Date       int    `json:"date"`
	Chat       Chat   `json:"chat"`
	Text       string `json:"text"`
}

type CallbackQuery struct {
	Id              string   `json:"id"`
	From            *User    `json:"form"`
	Message         *Message `json:"message"`
	InlineMessageId string   `json:"inline_message_id"`
	ChatInstance    string   `json:"chat_instance"`
	Data            string   `json:"data"`
	GameShortName   string   `json:"game_short_name"`
}

type Chat struct {
	Id            int      `json:"id"`
	Type          string   `json:"type"`
	Title         string   `json:"title,omitempty"`
	Username      string   `json:"username,omitempty"`
	FirstName     string   `json:"first_name,omitempty"`
	LastName      string   `json:"last_name,omitempty"`
	Bio           string   `json:"bio,omitempty"`
	Description   string   `json:"description,omitempty"`
	InviteLink    string   `json:"invite_link,omitempty"`
	PinnedMessage *Message `json:"pinned_message"`
	SlowModeDelay int      `json:"slow_mode_delay,omitempty"`
}

type SendMessageConfig struct {
	ParseMode                string
	DisableWebPagePreview    bool
	DisableNotification      bool
	ReplyToMessageId         int
	AllowSendingWithoutReply bool
	ReplyMarkup              *ReplyMarkup
}

type GetUpdatesConfig struct {
	Offset         int
	Limit          int
	Timeout        int
	AllowedUpdates []string
}

type ReplyMarkup struct {
	InlineKeyboardMarkup *InlineKeyboardMarkup
	ReplyKeyboardMarkup  *ReplyKeyboardMarkup
	ReplyKeyboardRemove  *ReplyKeyboardRemove
}

type InlineKeyboardMarkup struct {
	InlineKeyboard InlineKeyboard `json:"inline_keyboard"`
}

type InlineKeyboard []InlineKeyboardRow

type InlineKeyboardRow []InlineKeyboardButton

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	Url          string `json:"url,omitempty"`
	CallbackData string `json:"callback_data,omitempty"`
}

type ReplyKeyboardMarkup struct {
	Keyboard        ReplyKeyboard `json:"keyboard"`
	ResizeKeyboard  bool          `json:"resize_keyboard"`
	OneTimeKeyboard bool          `json:"one_time_keyboard"`
	Selective       bool          `json:"selective"`
}

type ReplyKeyboard []ReplyKeyboardRow

type ReplyKeyboardRow []ReplyKeyboardButton

type ReplyKeyboardButton struct {
	Text            string `json:"text"`
	RequestContact  bool   `json:"request_contact"`
	RequestLocation bool   `json:"request_location"`
}

type ReplyKeyboardRemove struct {
	RemoveKeyboard bool
	Selective      bool
}

// func NewSendMessageConfig()
