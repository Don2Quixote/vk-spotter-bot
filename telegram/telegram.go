package telegram

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func NewBot(token string) *Bot {
	bot := new(Bot)
	bot.token = token
	bot.apiEndpoint = "https://api.telegram.org/bot%s/%s"
	return bot
}

func (bot *Bot) call(methodName string, params url.Values) (*Response, error) {
	url := fmt.Sprintf(bot.apiEndpoint, bot.token, methodName)

	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	telegramResponse := new(Response)

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(telegramResponse)
	if err != nil {
		return nil, err
	}

	return telegramResponse, nil
}

func (bot *Bot) GetMe() (*User, error) {
	telegramResponse, err := bot.call("getMe", url.Values{})
	if err != nil {
		return nil, err
	}

	if !telegramResponse.Ok {
		if telegramResponse.Description != nil {
			return nil, errors.New(*telegramResponse.Description)
		}
		return nil, errors.New("No result in TelegramResponse")
	}

	user := new(User)
	err = json.Unmarshal(*telegramResponse.Result, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (bot *Bot) SendMessage(chatId int, text string, config *SendMessageConfig) (*Message, error) {
	params := url.Values{}

	params.Set("chat_id", strconv.Itoa(chatId))
	params.Set("text", text)
	if config != nil {
		if config.ParseMode != "" {
			params.Set("parse_mode", config.ParseMode)
		}
		if config.DisableWebPagePreview {
			params.Set("disable_web_page_preview", "true")
		}
		if config.DisableNotification {
			params.Set("disable_notification", "true")
		}
		if config.ReplyToMessageId != 0 {
			params.Set("reply_to_message_id", strconv.Itoa(config.ReplyToMessageId))
		}
		if config.AllowSendingWithoutReply {
			params.Set("allow_sending_without_reply", "true")
		}
		if config.ReplyMarkup != nil {
			if config.ReplyMarkup.InlineKeyboardMarkup != nil {
				jsonInlineKeyboardMarkup, err := json.Marshal(config.ReplyMarkup.InlineKeyboardMarkup)
				if err != nil {
					return nil, err
				}
				params.Set("reply_markup", string(jsonInlineKeyboardMarkup))
			} else if config.ReplyMarkup.ReplyKeyboardMarkup != nil {
				jsonReplyKeyboardMarkup, err := json.Marshal(config.ReplyMarkup.ReplyKeyboardMarkup)
				if err != nil {
					return nil, err
				}
				params.Set("reply_markup", string(jsonReplyKeyboardMarkup))
			} else if config.ReplyMarkup.ReplyKeyboardRemove != nil {
				jsonReplyKeyboardRemove, err := json.Marshal(config.ReplyMarkup.ReplyKeyboardRemove)
				if err != nil {
					return nil, err
				}
				params.Set("reply_markup", string(jsonReplyKeyboardRemove))
			}
		}
	}

	telegramResponse, err := bot.call("sendMessage", params)
	if err != nil {
		return nil, err
	}

	if !telegramResponse.Ok {
		if telegramResponse.Description != nil {
			return nil, errors.New(*telegramResponse.Description)
		}
		return nil, errors.New("No result in TelegramResponse")
	}

	message := new(Message)
	err = json.Unmarshal(*telegramResponse.Result, message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (bot *Bot) AnswerCallbackQuery(callbackQueryId string, text string, showAlert bool) (bool, error) {
	params := url.Values{}

	params.Set("callback_query_id", callbackQueryId)
	if text != "" {
		params.Set("text", text)
	}
	if showAlert {
		params.Set("show_alert", "true")
	}

	telegramResponse, err := bot.call("answerCallbackQuery", params)
	if err != nil {
		return false, err
	}

	if !telegramResponse.Ok {
		log.Println(telegramResponse)
		if telegramResponse.Description != nil {
			return false, errors.New(*telegramResponse.Description)
		} else {
			return false, errors.New("No result in TelegramResponse")
		}
	}

	var result bool
	err = json.Unmarshal(*telegramResponse.Result, &result)
	if err != nil {
		return false, err
	}

	return result, nil
}

func (bot *Bot) EditMessageReplyMarkup(chatId int, messageId int, replyMarkup *ReplyMarkup) (bool, error) {
	params := url.Values{}

	params.Set("chat_id", strconv.Itoa(chatId))
	params.Set("message_id", strconv.Itoa(messageId))
	if replyMarkup != nil && replyMarkup.InlineKeyboardMarkup != nil {
		jsonInlineKeyboardMarkup, err := json.Marshal(replyMarkup.InlineKeyboardMarkup)
		if err != nil {
			return false, err
		}
		params.Set("reply_markup", string(jsonInlineKeyboardMarkup))
	}

	telegramResponse, err := bot.call("editMessageReplyMarkup", params)
	if err != nil {
		return false, errors.New("No result in TelegramResponse")
	}

	if !telegramResponse.Ok {
		log.Println(telegramResponse)
		if telegramResponse.Description != nil {
			return false, errors.New(*telegramResponse.Description)
		} else {
			return false, errors.New("No result in TelegramResponse")
		}
	}

	return true, nil
}

func (bot *Bot) GetUpdates(config *GetUpdatesConfig) (*[]Update, error) {
	params := url.Values{}

	if config != nil {
		params.Set("offset", strconv.Itoa(config.Offset))
		if config.Limit != 0 {
			params.Set("limit", strconv.Itoa(config.Limit))
		}
		if config.Timeout != 0 {
			params.Set("timeout", strconv.Itoa(config.Timeout))
		}
		if config.AllowedUpdates != nil {
			jsonAllowedUpdates, err := json.Marshal(config.AllowedUpdates)
			if err != nil {
				return nil, err
			}
			params.Set("allowed_updates", string(jsonAllowedUpdates))
		}
	}

	telegramResponse, err := bot.call("getUpdates", params)
	if err != nil {
		return nil, err
	}

	if !telegramResponse.Ok {
		if telegramResponse.Description != nil {
			return nil, errors.New(*telegramResponse.Description)
		} else {
			return nil, errors.New("No result in TelegramResponse")
		}
	}

	updates := new([]Update)
	json.Unmarshal(*telegramResponse.Result, &updates)
	if err != nil {
		return nil, err
	}

	return updates, nil
}

func (bot *Bot) GrabUpdatesToChan(updatesChannel chan Update) {
	getUpdatesConfig := GetUpdatesConfig{
		Timeout: 30,
	}
	for {
		updates, err := bot.GetUpdates(&getUpdatesConfig)
		if err != nil {
			time.Sleep(time.Second * 5)
			continue
		}
		if len(*updates) == 0 {
			continue
		}
		var index int
		var update Update
		for index, update = range *updates {
			updatesChannel <- update
		}
		getUpdatesConfig.Offset = (*updates)[index].UpdateId + 1
	}
}
