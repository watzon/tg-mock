package faker

import (
	"time"
)

// registerGenerators registers all type-specific generators.
func (f *Faker) registerGenerators() {
	// Core types
	f.generators["User"] = (*Faker).generateUser
	f.generators["Chat"] = (*Faker).generateChat
	f.generators["ChatFullInfo"] = (*Faker).generateChatFullInfo
	f.generators["Message"] = (*Faker).generateMessage
	f.generators["MessageId"] = (*Faker).generateMessageId
	f.generators["File"] = (*Faker).generateFile
	f.generators["Update"] = (*Faker).generateUpdate

	// Media types
	f.generators["PhotoSize"] = (*Faker).generatePhotoSize
	f.generators["Audio"] = (*Faker).generateAudio
	f.generators["Document"] = (*Faker).generateDocument
	f.generators["Video"] = (*Faker).generateVideo
	f.generators["Animation"] = (*Faker).generateAnimation
	f.generators["Voice"] = (*Faker).generateVoice
	f.generators["VideoNote"] = (*Faker).generateVideoNote
	f.generators["Sticker"] = (*Faker).generateSticker
	f.generators["Contact"] = (*Faker).generateContact
	f.generators["Location"] = (*Faker).generateLocation
	f.generators["Venue"] = (*Faker).generateVenue
	f.generators["Poll"] = (*Faker).generatePoll
	f.generators["Dice"] = (*Faker).generateDice

	// Chat-related types
	f.generators["ChatMember"] = (*Faker).generateChatMember
	f.generators["ChatMemberOwner"] = (*Faker).generateChatMemberOwner
	f.generators["ChatMemberAdministrator"] = (*Faker).generateChatMemberAdministrator
	f.generators["ChatMemberMember"] = (*Faker).generateChatMemberMember
	f.generators["ChatInviteLink"] = (*Faker).generateChatInviteLink
	f.generators["ChatPhoto"] = (*Faker).generateChatPhoto
	f.generators["ChatPermissions"] = (*Faker).generateChatPermissions

	// Inline types
	f.generators["InlineQuery"] = (*Faker).generateInlineQuery
	f.generators["ChosenInlineResult"] = (*Faker).generateChosenInlineResult
	f.generators["CallbackQuery"] = (*Faker).generateCallbackQuery

	// Keyboard types
	f.generators["InlineKeyboardMarkup"] = (*Faker).generateInlineKeyboardMarkup
	f.generators["InlineKeyboardButton"] = (*Faker).generateInlineKeyboardButton
	f.generators["ReplyKeyboardMarkup"] = (*Faker).generateReplyKeyboardMarkup
	f.generators["KeyboardButton"] = (*Faker).generateKeyboardButton

	// Other types
	f.generators["WebhookInfo"] = (*Faker).generateWebhookInfo
	f.generators["BotCommand"] = (*Faker).generateBotCommand
	f.generators["BotDescription"] = (*Faker).generateBotDescription
	f.generators["BotName"] = (*Faker).generateBotName
	f.generators["BotShortDescription"] = (*Faker).generateBotShortDescription
	f.generators["MessageEntity"] = (*Faker).generateMessageEntity
	f.generators["UserProfilePhotos"] = (*Faker).generateUserProfilePhotos
	f.generators["ForumTopic"] = (*Faker).generateForumTopic
	f.generators["SentWebAppMessage"] = (*Faker).generateSentWebAppMessage
}

// Core type generators

func (f *Faker) generateUser(params map[string]interface{}) map[string]interface{} {
	userID := f.NextUserID() + 100000000

	// Check if user_id is provided in params
	if id, ok := params["user_id"].(float64); ok {
		userID = int64(id)
	} else if id, ok := params["user_id"].(int64); ok {
		userID = id
	}

	user := map[string]interface{}{
		"id":         userID,
		"is_bot":     false,
		"first_name": f.RandomChoice(firstNames),
	}

	// Add optional fields with some probability
	if f.RandomBool(0.7) {
		user["last_name"] = f.RandomChoice(lastNames)
	}
	if f.RandomBool(0.8) {
		user["username"] = f.generateUsername()
	}
	if f.RandomBool(0.5) {
		user["language_code"] = f.RandomChoice(languageCodes)
	}
	if f.RandomBool(0.3) {
		user["is_premium"] = true
	}

	return user
}

func (f *Faker) generateChat(params map[string]interface{}) map[string]interface{} {
	chatID := f.NextChatID()
	chatType := "private"

	// Check if chat_id is provided in params
	if id, ok := params["chat_id"].(float64); ok {
		chatID = int64(id)
	} else if id, ok := params["chat_id"].(int64); ok {
		chatID = id
	}

	// Negative IDs are typically groups/channels
	if chatID < 0 {
		if chatID < -1000000000000 {
			chatType = "channel"
		} else {
			chatType = f.RandomChoice([]string{"group", "supergroup"})
		}
	}

	chat := map[string]interface{}{
		"id":   chatID,
		"type": chatType,
	}

	// Add type-specific fields
	switch chatType {
	case "private":
		chat["first_name"] = f.RandomChoice(firstNames)
		if f.RandomBool(0.7) {
			chat["last_name"] = f.RandomChoice(lastNames)
		}
		if f.RandomBool(0.8) {
			chat["username"] = f.generateUsername()
		}
	case "group", "supergroup":
		chat["title"] = f.generateTitle()
		if f.RandomBool(0.6) {
			chat["username"] = f.generateUsername()
		}
	case "channel":
		chat["title"] = f.generateTitle()
		if f.RandomBool(0.8) {
			chat["username"] = f.generateUsername()
		}
	}

	return chat
}

func (f *Faker) generateChatFullInfo(params map[string]interface{}) map[string]interface{} {
	// Start with basic chat info
	chat := f.generateChat(params)

	// Add full info fields
	chat["accent_color_id"] = f.RandomInt64(0, 20)
	chat["max_reaction_count"] = int64(11)

	if f.RandomBool(0.6) {
		chat["photo"] = f.generateChatPhoto(params)
	}
	if f.RandomBool(0.5) {
		chat["bio"] = f.generateText()
	}
	if f.RandomBool(0.4) {
		chat["description"] = f.generateText()
	}

	return chat
}

func (f *Faker) generateMessage(params map[string]interface{}) map[string]interface{} {
	messageID := f.NextMessageID()
	chatID := int64(1)

	// Extract chat_id from params
	if id, ok := params["chat_id"].(float64); ok {
		chatID = int64(id)
	} else if id, ok := params["chat_id"].(int64); ok {
		chatID = id
	}

	msg := map[string]interface{}{
		"message_id": messageID,
		"date":       time.Now().Unix(),
		"chat":       f.generateChat(map[string]interface{}{"chat_id": chatID}),
	}

	// Add from user for non-channel messages
	chatType := msg["chat"].(map[string]interface{})["type"].(string)
	if chatType != "channel" {
		msg["from"] = f.generateUser(params)
	}

	// Reflect text from params
	if text, ok := params["text"].(string); ok {
		msg["text"] = text
	}

	// Reflect caption from params
	if caption, ok := params["caption"].(string); ok {
		msg["caption"] = caption
	}

	// Handle media types based on method context
	if _, ok := params["photo"]; ok {
		msg["photo"] = f.generatePhotoSizes()
	}
	if _, ok := params["document"]; ok {
		msg["document"] = f.generateDocument(params)
	}
	if _, ok := params["audio"]; ok {
		msg["audio"] = f.generateAudio(params)
	}
	if _, ok := params["video"]; ok {
		msg["video"] = f.generateVideo(params)
	}
	if _, ok := params["voice"]; ok {
		msg["voice"] = f.generateVoice(params)
	}
	if _, ok := params["animation"]; ok {
		msg["animation"] = f.generateAnimation(params)
	}
	if _, ok := params["sticker"]; ok {
		msg["sticker"] = f.generateSticker(params)
	}
	if _, ok := params["location"]; ok || params["latitude"] != nil {
		msg["location"] = f.generateLocation(params)
	}
	if _, ok := params["venue"]; ok {
		msg["venue"] = f.generateVenue(params)
	}
	if _, ok := params["contact"]; ok {
		msg["contact"] = f.generateContact(params)
	}
	if _, ok := params["poll"]; ok {
		msg["poll"] = f.generatePoll(params)
	}
	if _, ok := params["dice"]; ok {
		msg["dice"] = f.generateDice(params)
	}

	// Handle reply markup
	if _, ok := params["reply_markup"]; ok {
		// Reply markup is passed through but not generated
	}

	return msg
}

func (f *Faker) generateMessageId(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"message_id": f.NextMessageID(),
	}
}

func (f *Faker) generateFile(params map[string]interface{}) map[string]interface{} {
	fileID := f.generateFileID()
	if id, ok := params["file_id"].(string); ok {
		fileID = id
	}

	uniqueID := "unique_"
	if len(fileID) >= 8 {
		uniqueID += fileID[:8]
	} else {
		uniqueID += fileID
	}

	return map[string]interface{}{
		"file_id":        fileID,
		"file_unique_id": uniqueID,
		"file_size":      f.RandomInt64(1024, 1024*1024*10),
		"file_path":      f.generateFilePath(),
	}
}

func (f *Faker) generateUpdate(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"update_id": f.NextUpdateID(),
	}
}

// Media type generators

func (f *Faker) generatePhotoSize(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"file_id":        f.generateFileID(),
		"file_unique_id": f.generateFileID()[:20],
		"width":          f.RandomInt64(100, 1920),
		"height":         f.RandomInt64(100, 1080),
		"file_size":      f.RandomInt64(1024, 1024*500),
	}
}

func (f *Faker) generatePhotoSizes() []map[string]interface{} {
	// Generate 3 sizes: small, medium, large
	sizes := []struct {
		w, h int64
	}{
		{90, 90},
		{320, 320},
		{800, 800},
	}

	result := make([]map[string]interface{}, len(sizes))
	for i, size := range sizes {
		result[i] = map[string]interface{}{
			"file_id":        f.generateFileID(),
			"file_unique_id": f.generateFileID()[:20],
			"width":          size.w,
			"height":         size.h,
			"file_size":      f.RandomInt64(1024, 1024*100*(int64(i)+1)),
		}
	}
	return result
}

func (f *Faker) generateAudio(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"file_id":        f.generateFileID(),
		"file_unique_id": f.generateFileID()[:20],
		"duration":       f.RandomInt64(30, 300),
		"performer":      f.generateAuthor(),
		"title":          f.generateTitle(),
		"mime_type":      "audio/mpeg",
		"file_size":      f.RandomInt64(1024*100, 1024*1024*10),
	}
}

func (f *Faker) generateDocument(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"file_id":        f.generateFileID(),
		"file_unique_id": f.generateFileID()[:20],
		"file_name":      f.generateFilePath(),
		"mime_type":      f.RandomChoice(mimeTypes),
		"file_size":      f.RandomInt64(1024, 1024*1024*50),
	}
}

func (f *Faker) generateVideo(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"file_id":        f.generateFileID(),
		"file_unique_id": f.generateFileID()[:20],
		"width":          f.RandomInt64(320, 1920),
		"height":         f.RandomInt64(240, 1080),
		"duration":       f.RandomInt64(5, 600),
		"mime_type":      "video/mp4",
		"file_size":      f.RandomInt64(1024*100, 1024*1024*100),
	}
}

func (f *Faker) generateAnimation(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"file_id":        f.generateFileID(),
		"file_unique_id": f.generateFileID()[:20],
		"width":          f.RandomInt64(100, 500),
		"height":         f.RandomInt64(100, 500),
		"duration":       f.RandomInt64(1, 10),
		"mime_type":      "video/mp4",
		"file_size":      f.RandomInt64(1024*10, 1024*1024*5),
	}
}

func (f *Faker) generateVoice(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"file_id":        f.generateFileID(),
		"file_unique_id": f.generateFileID()[:20],
		"duration":       f.RandomInt64(1, 120),
		"mime_type":      "audio/ogg",
		"file_size":      f.RandomInt64(1024, 1024*1024),
	}
}

func (f *Faker) generateVideoNote(params map[string]interface{}) map[string]interface{} {
	length := f.RandomInt64(200, 500)
	return map[string]interface{}{
		"file_id":        f.generateFileID(),
		"file_unique_id": f.generateFileID()[:20],
		"length":         length,
		"duration":       f.RandomInt64(1, 60),
		"file_size":      f.RandomInt64(1024*100, 1024*1024*10),
	}
}

func (f *Faker) generateSticker(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"file_id":        f.generateFileID(),
		"file_unique_id": f.generateFileID()[:20],
		"type":           f.RandomChoice([]string{"regular", "mask", "custom_emoji"}),
		"width":          int64(512),
		"height":         int64(512),
		"is_animated":    f.RandomBool(0.3),
		"is_video":       f.RandomBool(0.2),
		"file_size":      f.RandomInt64(1024*10, 1024*100),
	}
}

func (f *Faker) generateContact(params map[string]interface{}) map[string]interface{} {
	contact := map[string]interface{}{
		"phone_number": f.generatePhoneNumber(),
		"first_name":   f.RandomChoice(firstNames),
	}
	if f.RandomBool(0.7) {
		contact["last_name"] = f.RandomChoice(lastNames)
	}
	if f.RandomBool(0.5) {
		contact["user_id"] = f.RandomInt64(100000000, 999999999)
	}
	return contact
}

func (f *Faker) generateLocation(params map[string]interface{}) map[string]interface{} {
	loc := map[string]interface{}{
		"latitude":  f.RandomFloat64(-90, 90),
		"longitude": f.RandomFloat64(-180, 180),
	}

	// Use provided coordinates if available
	if lat, ok := params["latitude"].(float64); ok {
		loc["latitude"] = lat
	}
	if lon, ok := params["longitude"].(float64); ok {
		loc["longitude"] = lon
	}

	if f.RandomBool(0.3) {
		loc["horizontal_accuracy"] = f.RandomFloat64(0, 100)
	}

	return loc
}

func (f *Faker) generateVenue(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"location":        f.generateLocation(params),
		"title":           f.generateTitle(),
		"address":         f.generateText(),
		"foursquare_id":   f.generateFileID()[:24],
		"foursquare_type": "food/restaurant",
	}
}

func (f *Faker) generatePoll(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"id":                    f.generateFileID()[:17],
		"question":              f.generateText(),
		"options":               []map[string]interface{}{},
		"total_voter_count":     f.RandomInt64(0, 100),
		"is_closed":             false,
		"is_anonymous":          true,
		"type":                  "regular",
		"allows_multiple_answers": false,
	}
}

func (f *Faker) generateDice(params map[string]interface{}) map[string]interface{} {
	emoji := "🎲"
	if e, ok := params["emoji"].(string); ok {
		emoji = e
	}

	maxValue := int64(6)
	switch emoji {
	case "🎯":
		maxValue = 6
	case "🏀":
		maxValue = 5
	case "⚽":
		maxValue = 5
	case "🎰":
		maxValue = 64
	case "🎳":
		maxValue = 6
	}

	return map[string]interface{}{
		"emoji": emoji,
		"value": f.RandomInt64(1, maxValue+1),
	}
}

// Chat member generators

func (f *Faker) generateChatMember(params map[string]interface{}) map[string]interface{} {
	// Default to regular member
	return f.generateChatMemberMember(params)
}

func (f *Faker) generateChatMemberOwner(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"status":       "creator",
		"user":         f.generateUser(params),
		"is_anonymous": f.RandomBool(0.2),
	}
}

func (f *Faker) generateChatMemberAdministrator(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"status":                "administrator",
		"user":                  f.generateUser(params),
		"can_be_edited":         true,
		"is_anonymous":          f.RandomBool(0.2),
		"can_manage_chat":       true,
		"can_delete_messages":   true,
		"can_manage_video_chats": true,
		"can_restrict_members":  true,
		"can_promote_members":   f.RandomBool(0.5),
		"can_change_info":       true,
		"can_invite_users":      true,
		"can_post_messages":     true,
		"can_edit_messages":     true,
		"can_pin_messages":      true,
	}
}

func (f *Faker) generateChatMemberMember(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"status": "member",
		"user":   f.generateUser(params),
	}
}

func (f *Faker) generateChatInviteLink(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"invite_link":          "https://t.me/+" + f.generateFileID()[:16],
		"creator":              f.generateUser(params),
		"creates_join_request": f.RandomBool(0.3),
		"is_primary":           f.RandomBool(0.5),
		"is_revoked":           false,
	}
}

func (f *Faker) generateChatPhoto(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"small_file_id":        f.generateFileID(),
		"small_file_unique_id": f.generateFileID()[:20],
		"big_file_id":          f.generateFileID(),
		"big_file_unique_id":   f.generateFileID()[:20],
	}
}

func (f *Faker) generateChatPermissions(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"can_send_messages":         true,
		"can_send_audios":           true,
		"can_send_documents":        true,
		"can_send_photos":           true,
		"can_send_videos":           true,
		"can_send_video_notes":      true,
		"can_send_voice_notes":      true,
		"can_send_polls":            true,
		"can_send_other_messages":   true,
		"can_add_web_page_previews": true,
		"can_change_info":           f.RandomBool(0.5),
		"can_invite_users":          true,
		"can_pin_messages":          f.RandomBool(0.5),
	}
}

// Inline type generators

func (f *Faker) generateInlineQuery(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"id":     f.generateFileID()[:20],
		"from":   f.generateUser(params),
		"query":  f.generateQuery(),
		"offset": "",
	}
}

func (f *Faker) generateChosenInlineResult(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"result_id":         f.generateFileID()[:20],
		"from":              f.generateUser(params),
		"query":             f.generateQuery(),
		"inline_message_id": f.generateFileID()[:30],
	}
}

func (f *Faker) generateCallbackQuery(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"id":            f.generateFileID()[:20],
		"from":          f.generateUser(params),
		"chat_instance": f.generateFileID()[:15],
		"data":          f.generateString("callback_data"),
	}
}

// Keyboard generators

func (f *Faker) generateInlineKeyboardMarkup(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{},
	}
}

func (f *Faker) generateInlineKeyboardButton(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"text":          "Button",
		"callback_data": f.generateString("callback_data"),
	}
}

func (f *Faker) generateReplyKeyboardMarkup(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"keyboard":        [][]map[string]interface{}{},
		"resize_keyboard": true,
	}
}

func (f *Faker) generateKeyboardButton(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"text": "Button",
	}
}

// Other generators

func (f *Faker) generateWebhookInfo(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"url":                    "",
		"has_custom_certificate": false,
		"pending_update_count":   int64(0),
	}
}

func (f *Faker) generateBotCommand(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"command":     f.RandomChoice(commands),
		"description": f.generateText(),
	}
}

func (f *Faker) generateBotDescription(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"description": f.generateText(),
	}
}

func (f *Faker) generateBotName(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name": f.generateTitle(),
	}
}

func (f *Faker) generateBotShortDescription(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"short_description": f.generateText(),
	}
}

func (f *Faker) generateMessageEntity(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type":   f.RandomChoice([]string{"bold", "italic", "code", "mention", "hashtag", "url"}),
		"offset": int64(0),
		"length": f.RandomInt64(1, 20),
	}
}

func (f *Faker) generateUserProfilePhotos(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"total_count": int64(1),
		"photos":      [][]map[string]interface{}{f.generatePhotoSizes()},
	}
}

func (f *Faker) generateForumTopic(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"message_thread_id": f.RandomInt64(1, 10000),
		"name":              f.generateTitle(),
		"icon_color":        f.RandomInt64(0, 16777215),
	}
}

func (f *Faker) generateSentWebAppMessage(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"inline_message_id": f.generateFileID()[:30],
	}
}
