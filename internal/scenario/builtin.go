// Package scenario provides pre-built error responses for common Telegram API errors.
package scenario

// BuiltinErrors contains all common Telegram API error responses.
// These can be triggered via the X-TG-Mock-Scenario header.
var BuiltinErrors = map[string]*ErrorResponse{
	// 400 Bad Request - General
	"bad_request": {ErrorCode: 400, Description: "Bad Request"},

	// 400 Bad Request - Chat errors
	"chat_not_found":            {ErrorCode: 400, Description: "Bad Request: chat not found"},
	"chat_admin_required":       {ErrorCode: 400, Description: "Bad Request: CHAT_ADMIN_REQUIRED"},
	"chat_not_modified":         {ErrorCode: 400, Description: "Bad Request: CHAT_NOT_MODIFIED"},
	"chat_restricted":           {ErrorCode: 400, Description: "Bad Request: CHAT_RESTRICTED"},
	"chat_write_forbidden":      {ErrorCode: 400, Description: "Bad Request: CHAT_WRITE_FORBIDDEN"},
	"channel_private":           {ErrorCode: 400, Description: "Bad Request: CHANNEL_PRIVATE"},
	"group_deactivated":         {ErrorCode: 400, Description: "Bad Request: group is deactivated"},
	"group_upgraded":            {ErrorCode: 400, Description: "Bad Request: group chat was upgraded to a supergroup chat"},
	"supergroup_channel_only":   {ErrorCode: 400, Description: "Bad Request: method is available for supergroup and channel chats only"},
	"not_in_chat":               {ErrorCode: 400, Description: "Bad Request: not in the chat"},
	"topic_not_modified":        {ErrorCode: 400, Description: "Bad Request: TOPIC_NOT_MODIFIED"},

	// 400 Bad Request - User errors
	"user_not_found":            {ErrorCode: 400, Description: "Bad Request: user not found"},
	"user_id_invalid":           {ErrorCode: 400, Description: "Bad Request: USER_ID_INVALID"},
	"user_is_admin":             {ErrorCode: 400, Description: "Bad Request: user is an administrator of the chat"},
	"participant_id_invalid":    {ErrorCode: 400, Description: "Bad Request: PARTICIPANT_ID_INVALID"},
	"cant_remove_owner":         {ErrorCode: 400, Description: "Bad Request: can't remove chat owner"},

	// 400 Bad Request - Message errors
	"message_not_found":         {ErrorCode: 400, Description: "Bad Request: message to edit not found"},
	"message_not_modified":      {ErrorCode: 400, Description: "Bad Request: message is not modified"},
	"message_text_empty":        {ErrorCode: 400, Description: "Bad Request: message text is empty"},
	"message_too_long":          {ErrorCode: 400, Description: "Bad Request: message is too long"},
	"message_cant_be_edited":    {ErrorCode: 400, Description: "Bad Request: message can't be edited"},
	"message_cant_be_deleted":   {ErrorCode: 400, Description: "Bad Request: message can't be deleted"},
	"message_to_delete_not_found": {ErrorCode: 400, Description: "Bad Request: message to delete not found"},
	"message_id_invalid":        {ErrorCode: 400, Description: "Bad Request: MESSAGE_ID_INVALID"},
	"message_thread_not_found":  {ErrorCode: 400, Description: "Bad Request: message thread not found"},
	"reply_message_not_found":   {ErrorCode: 400, Description: "Bad Request: reply message not found"},

	// 400 Bad Request - Permission/Rights errors
	"no_rights_to_send":         {ErrorCode: 400, Description: "Bad Request: have no rights to send a message"},
	"not_enough_rights":         {ErrorCode: 400, Description: "Bad Request: not enough rights"},
	"not_enough_rights_pin":     {ErrorCode: 400, Description: "Bad Request: not enough rights to manage pinned messages in the chat"},
	"not_enough_rights_restrict": {ErrorCode: 400, Description: "Bad Request: not enough rights to restrict/unrestrict chat member"},
	"not_enough_rights_send_text": {ErrorCode: 400, Description: "Bad Request: not enough rights to send text messages to the chat"},

	// 400 Bad Request - Admin errors
	"admin_rank_emoji_not_allowed": {ErrorCode: 400, Description: "Bad Request: ADMIN_RANK_EMOJI_NOT_ALLOWED"},

	// 400 Bad Request - Inline/Button errors
	"button_url_invalid":        {ErrorCode: 400, Description: "Bad Request: BUTTON_URL_INVALID"},
	"inline_button_url_invalid": {ErrorCode: 400, Description: "Bad Request: inline keyboard button URL"},

	// 400 Bad Request - File errors
	"file_too_big":              {ErrorCode: 400, Description: "Bad Request: file is too big"},
	"invalid_file_id":           {ErrorCode: 400, Description: "Bad Request: invalid file id"},

	// 400 Bad Request - Other
	"entities_too_long":         {ErrorCode: 400, Description: "Bad Request: entities too long"},
	"member_not_found":          {ErrorCode: 400, Description: "Bad Request: member not found"},
	"peer_id_invalid":           {ErrorCode: 400, Description: "Bad Request: PEER_ID_INVALID"},
	"wrong_parameter_action":    {ErrorCode: 400, Description: "Bad Request: wrong parameter action in request"},
	"hide_requester_missing":    {ErrorCode: 400, Description: "Bad Request: HIDE_REQUESTER_MISSING"},

	// 401 Unauthorized
	"unauthorized": {ErrorCode: 401, Description: "Unauthorized"},

	// 403 Forbidden - General
	"forbidden":               {ErrorCode: 403, Description: "Forbidden"},

	// 403 Forbidden - Bot blocked/kicked
	"bot_blocked":             {ErrorCode: 403, Description: "Forbidden: bot was blocked by the user"},
	"bot_kicked":              {ErrorCode: 403, Description: "Forbidden: bot was kicked from the chat"},
	"bot_kicked_channel":      {ErrorCode: 403, Description: "Forbidden: bot was kicked from the channel chat"},
	"bot_kicked_group":        {ErrorCode: 403, Description: "Forbidden: bot was kicked from the group chat"},
	"bot_kicked_supergroup":   {ErrorCode: 403, Description: "Forbidden: bot was kicked from the supergroup chat"},

	// 403 Forbidden - Bot not member
	"not_member_channel":      {ErrorCode: 403, Description: "Forbidden: bot is not a member of the channel chat"},
	"not_member_supergroup":   {ErrorCode: 403, Description: "Forbidden: bot is not a member of the supergroup chat"},

	// 403 Forbidden - Bot can't act
	"cant_initiate":           {ErrorCode: 403, Description: "Forbidden: bot can't initiate conversation with a user"},
	"cant_send_to_bots":       {ErrorCode: 403, Description: "Forbidden: bot can't send messages to bots"},

	// 403 Forbidden - User status
	"user_deactivated":        {ErrorCode: 403, Description: "Forbidden: user is deactivated"},

	// 403 Forbidden - Permissions
	"not_enough_rights_text":  {ErrorCode: 403, Description: "Forbidden: not enough rights to send text messages"},
	"not_enough_rights_photo": {ErrorCode: 403, Description: "Forbidden: not enough rights to send photos"},

	// 409 Conflict
	"webhook_active":          {ErrorCode: 409, Description: "Conflict: can't use getUpdates method while webhook is active"},
	"terminated_by_long_poll": {ErrorCode: 409, Description: "Conflict: terminated by other long poll"},

	// 429 Rate Limit
	"rate_limit": {ErrorCode: 429, Description: "Too Many Requests: retry after 30", RetryAfter: 30},
	"flood_wait": {ErrorCode: 429, Description: "Flood control exceeded. Retry in 60 seconds", RetryAfter: 60},
}

// GetBuiltinError returns the pre-built error response for the given error name.
// Returns nil if the error name is not found.
func GetBuiltinError(name string) *ErrorResponse {
	return BuiltinErrors[name]
}
