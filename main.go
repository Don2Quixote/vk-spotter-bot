package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"./telegram"
)

type Target struct {
	Id              int
	Domain          string
	DomainIsPrimary bool
	FirstName       string
	LastName        string
	LastSeenTime    int
}

type Targets []*Target

func (targets *Targets) find(id int) *Target {
	for _, target := range *targets {
		if target.Id == id {
			return target
		}
	}
	return nil
}

func (targets *Targets) add(target *Target) {
	*targets = append(*targets, target)
}

func (targets *Targets) remove(id int) {
	newTargets := make(Targets, 0)
	for _, target := range *targets {
		if target.Id != id {
			newTargets = append(newTargets, target)
		}
	}
	*targets = newTargets
}

func (targets *Targets) clear() {
	*targets = []*Target{}
}

var targets Targets

func (targets *Targets) startTracing(bot *telegram.Bot) {
	for {
		time.Sleep(time.Second * 7)
		userIdsToGet := []string{}
		for _, target := range *targets {
			userIdsToGet = append(userIdsToGet, strconv.Itoa(target.Id))
		}
		users, err := vkGetUsers(userIdsToGet)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		for _, user := range users {
			for _, target := range *targets {
				if user.Id == target.Id {
					if user.Online == 1 || user.LastSeen.Time != target.LastSeenTime {
						targets.remove(target.Id)

						var notificationMessageText string
						if target.DomainIsPrimary {
							notificationMessageText = fmt.Sprintf("✉️ %s (%s %s) Online", target.Domain, target.FirstName, target.LastName)
						} else {
							notificationMessageText = fmt.Sprintf("✉️ %d (%s %s) Online", target.Id, target.FirstName, target.LastName)
						}
						var domainIsPrimaryCallbackArg string
						if target.DomainIsPrimary {
							domainIsPrimaryCallbackArg = "true"
						} else {
							domainIsPrimaryCallbackArg = "false"
						}
						bot.SendMessage(OWNER_ID, notificationMessageText, &telegram.SendMessageConfig{
							ReplyMarkup: &telegram.ReplyMarkup{
								InlineKeyboardMarkup: &telegram.InlineKeyboardMarkup{
									InlineKeyboard: telegram.InlineKeyboard{
										telegram.InlineKeyboardRow{
											telegram.InlineKeyboardButton{
												Text: "🔄 Repeat", CallbackData: fmt.Sprintf("repeat:%d:%s", target.Id, domainIsPrimaryCallbackArg),
											},
										},
									},
								},
							},
						})
					}
				}
			}
		}
	}
}

var VK_TOKEN, TG_TOKEN, OWNER_ID = "", "", 0

type vkUser struct {
	Id              int    `json:"id"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	IsClosed        bool   `json:"is_closed"`
	CanAccessClosed bool   `json:"can_access_closed"`
	Domain          string `json:"domain"`
	Online          int    `json:"online"`
	LastSeen        struct {
		Platfrom int `json:"platform"`
		Time     int `json:"time"`
	} `json:"last_seen"`
}

func vkGetUsers(userIds []string) ([]vkUser, error) {
	params := url.Values{}
	params.Set("access_token", VK_TOKEN)
	params.Set("v", "5.126")
	params.Set("lang", "ru")
	params.Set("user_ids", strings.Join(userIds, ","))
	params.Set("fields", "last_seen,online,domain")

	url := "https://api.vk.com/method/users.get?" + params.Encode()
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	var responseStruct struct {
		Response *[]vkUser `json:"response"`
	}

	decoder := json.NewDecoder(response.Body)
	decoder.Decode(&responseStruct)

	if responseStruct.Response == nil {
		return nil, errors.New("users.get method error")
	}

	return *responseStruct.Response, nil
}

func handleMessage(bot *telegram.Bot, message *telegram.Message) {
	splittedMessage := strings.Split(message.Text, " ")
	command := splittedMessage[0]
	var args []string
	if len(splittedMessage) > 1 {
		args = splittedMessage[1:]
	}

	if command == "/start" {
		bot.SendMessage(OWNER_ID, "👋 Hello", &telegram.SendMessageConfig{
			ReplyMarkup: &telegram.ReplyMarkup{
				ReplyKeyboardMarkup: &telegram.ReplyKeyboardMarkup{
					Keyboard: telegram.ReplyKeyboard{
						telegram.ReplyKeyboardRow{
							telegram.ReplyKeyboardButton{Text: "📝 List"},
						},
						telegram.ReplyKeyboardRow{
							telegram.ReplyKeyboardButton{Text: "♻️ Clear List"},
						},
					},
					ResizeKeyboard: true,
				},
			},
		})
	} else if command == "/add" {
		if len(args) == 0 {
			bot.SendMessage(OWNER_ID, "ℹ️ No arguments", nil)
			return
		}

		userIdsToGet := []string{}
		for _, vkIdOrDomain := range args {
			found := false
			for _, userIdToGet := range userIdsToGet {
				if userIdToGet == vkIdOrDomain {
					found = true
					break
				}
			}
			if found {
				continue
			}
			if vkIdOrDomain != "" {
				userIdsToGet = append(userIdsToGet, vkIdOrDomain)
			}
		}

		users, err := vkGetUsers(userIdsToGet)
		if err != nil {
			log.Println(err.Error())
			return
		}

		sendingMessages := sync.WaitGroup{}
		sendingMessages.Add(len(userIdsToGet))

		for _, user := range users {
			var domainIsPrimary bool
			for i, id := range userIdsToGet {
				if id == user.Domain {
					domainIsPrimary = true
					userIdsToGet[i] = ""
				} else if id == strconv.Itoa(user.Id) {
					domainIsPrimary = false
					userIdsToGet[i] = ""
				}
			}

			if targets.find(user.Id) != nil {
				var replyText string
				if domainIsPrimary {
					replyText = fmt.Sprintf("ℹ️ %s (%s %s) Already added", user.Domain, user.FirstName, user.LastName)
				} else {
					replyText = fmt.Sprintf("ℹ️ %d (%s %s) Already added", user.Id, user.FirstName, user.LastName)
				}
				go func() {
					bot.SendMessage(OWNER_ID, replyText, nil)
					sendingMessages.Done()
				}()
				continue
			}

			if user.Online == 1 {
				var replyText string
				if domainIsPrimary {
					replyText = fmt.Sprintf("✉️ %s (%s %s) Online", user.Domain, user.FirstName, user.LastName)
				} else {
					replyText = fmt.Sprintf("✉️ %d (%s %s) Online", user.Id, user.FirstName, user.LastName)
				}
				go func() {
					bot.SendMessage(OWNER_ID, replyText, nil)
					sendingMessages.Done()
				}()
				continue
			}

			targets.add(&Target{
				Id:              user.Id,
				Domain:          user.Domain,
				DomainIsPrimary: domainIsPrimary,
				FirstName:       user.FirstName,
				LastName:        user.LastName,
				LastSeenTime:    user.LastSeen.Time,
			})

			var replyText string
			if domainIsPrimary {
				replyText = fmt.Sprintf("✅ %s (%s %s) Added", user.Domain, user.FirstName, user.LastName)
			} else {
				replyText = fmt.Sprintf("✅ %d (%s %s) Added", user.Id, user.FirstName, user.LastName)
			}
			go func() {
				bot.SendMessage(OWNER_ID, replyText, nil)
				sendingMessages.Done()
			}()
		}
		for _, id := range userIdsToGet {
			if id != "" {
				replyText := fmt.Sprintf("❌ %s Not found", id)
				go func() {
					bot.SendMessage(OWNER_ID, replyText, nil)
					sendingMessages.Done()
				}()
			}
		}

		sendingMessages.Wait()
	} else if command == "/remove" {
		if len(args) == 0 {
			bot.SendMessage(OWNER_ID, "ℹ️ No arguments", nil)
			return
		}

		for _, vkIdOrDomain := range args {
			found := false
			for _, target := range targets {
				if target.Domain == vkIdOrDomain || strconv.Itoa(target.Id) == vkIdOrDomain {
					found = true
					targets.remove(target.Id)
					var replyText string
					if target.DomainIsPrimary {
						replyText = fmt.Sprintf("✅ %s (%s %s) Removed", target.Domain, target.FirstName, target.LastName)
					} else {
						replyText = fmt.Sprintf("✅ %d (%s %s) Removed", target.Id, target.FirstName, target.LastName)
					}
					bot.SendMessage(OWNER_ID, replyText, nil)
					break
				}
			}
			if !found {
				bot.SendMessage(OWNER_ID, fmt.Sprintf("❌ %s Not found in tracing list", vkIdOrDomain), nil)
			}
		}
	} else if command == "/clear" || command == "♻️" {
		if len(targets) == 0 {
			bot.SendMessage(OWNER_ID, fmt.Sprintf("ℹ️ Tracing list is empty"), nil)
			return
		}

		targets.clear()
		bot.SendMessage(OWNER_ID, "✅ Tracing list cleared", nil)
	} else if command == "/list" || command == "📝" {
		replyText := "📝 Tracing list"
		if len(targets) == 0 {
			replyText += " is empty"
		} else {
			replyText += "\n\n"
		}
		for i, target := range targets {
			if target.DomainIsPrimary {
				replyText += fmt.Sprintf("%d. %s (%s %s)\n", i+1, target.Domain, target.FirstName, target.LastName)
			} else {
				replyText += fmt.Sprintf("%d. %d (%s %s)\n", i+1, target.Id, target.FirstName, target.LastName)
			}
		}
		bot.SendMessage(OWNER_ID, replyText, nil)
	} else {
		bot.SendMessage(OWNER_ID, "ℹ️ Unknown command", nil)
	}
}

func handleCallback(bot *telegram.Bot, callback *telegram.CallbackQuery) {
	splittedCallbackData := strings.Split(callback.Data, ":")
	command := splittedCallbackData[0]
	var args []string
	if len(splittedCallbackData) > 1 {
		args = splittedCallbackData[1:]
	}

	if command == "repeat" {
		if len(args) != 2 {
			bot.AnswerCallbackQuery(callback.Id, "❌ Error occurred", false)
			return
		}

		var domainIsPrimary bool
		if args[1] == "true" {
			domainIsPrimary = true
		} else {
			domainIsPrimary = false
		}

		users, err := vkGetUsers(args[0:1])
		if err != nil {
			bot.AnswerCallbackQuery(callback.Id, "❌ Error occurred", false)
			return
		}

		if len(users) != 1 {
			bot.AnswerCallbackQuery(callback.Id, "❌ Error occurred", false)
			return
		}

		user := users[0]

		if user.Online == 1 {
			bot.AnswerCallbackQuery(callback.Id, "ℹ️ User is online", false)
			return
		}

		targets.add(&Target{
			Id:              user.Id,
			Domain:          user.Domain,
			DomainIsPrimary: domainIsPrimary,
			FirstName:       user.FirstName,
			LastName:        user.LastName,
			LastSeenTime:    user.LastSeen.Time,
		})

		bot.AnswerCallbackQuery(callback.Id, "✅ User added again", false)
		bot.EditMessageReplyMarkup(callback.Message.Chat.Id, callback.Message.MessageId, &telegram.ReplyMarkup{
			InlineKeyboardMarkup: &telegram.InlineKeyboardMarkup{
				InlineKeyboard: telegram.InlineKeyboard{
					telegram.InlineKeyboardRow{
						telegram.InlineKeyboardButton{
							Text: "ℹ️ Repeated", CallbackData: "plug",
						},
					},
				},
			},
		})
	} else {
		bot.AnswerCallbackQuery(callback.Id, "", false)
	}
}

func main() {
	TG_TOKEN = os.Getenv("TG_TOKEN")
	VK_TOKEN = os.Getenv("VK_TOKEN")
	ownerIdString := os.Getenv("OWNER_ID")
	if TG_TOKEN == "" {
		fmt.Println("TG_TOKEN Not specified")
		return
	}
	if VK_TOKEN == "" {
		fmt.Println("VK_TOKEN Not specified")
		return
	}
	if ownerIdString == "" {
		fmt.Println("OWNER_ID Not specified")
		return
	}
	OWNER_ID, err := strconv.Atoi(ownerIdString)
	if err != nil {
		fmt.Println("OWNER_ID Must be a number")
		return
	}

	bot := telegram.NewBot(TG_TOKEN)

	go targets.startTracing(bot)

	updates := make(chan telegram.Update)
	go bot.GrabUpdatesToChan(updates)

	for update := range updates {
		if update.Message != nil {
			if update.Message.From.Id == OWNER_ID && update.Message.Chat.Type == "private" {
				go handleMessage(bot, update.Message)
			}
		} else if update.CallbackQuery != nil {
			go handleCallback(bot, update.CallbackQuery)
		}
	}
}
