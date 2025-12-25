// Package scenario provides pre-built error responses for common Telegram API errors.
package scenario

// BuiltinErrors contains all common Telegram API error responses.
// These can be triggered via the X-TG-Mock-Scenario header.
var BuiltinErrors = map[string]*ErrorResponse{
	// 400 Bad Request
	"bad_request":             {ErrorCode: 400, Description: "Bad Request"},
	"chat_not_found":          {ErrorCode: 400, Description: "Bad Request: chat not found"},
	"user_not_found":          {ErrorCode: 400, Description: "Bad Request: user not found"},
	"message_not_found":       {ErrorCode: 400, Description: "Bad Request: message to edit not found"},
	"message_not_modified":    {ErrorCode: 400, Description: "Bad Request: message is not modified"},
	"message_text_empty":      {ErrorCode: 400, Description: "Bad Request: message text is empty"},
	"message_too_long":        {ErrorCode: 400, Description: "Bad Request: message is too long"},
	"message_cant_be_edited":  {ErrorCode: 400, Description: "Bad Request: message can't be edited"},
	"message_cant_be_deleted": {ErrorCode: 400, Description: "Bad Request: message can't be deleted"},
	"reply_message_not_found": {ErrorCode: 400, Description: "Bad Request: reply message not found"},
	"button_url_invalid":      {ErrorCode: 400, Description: "Bad Request: BUTTON_URL_INVALID"},
	"entities_too_long":       {ErrorCode: 400, Description: "Bad Request: entities too long"},
	"file_too_big":            {ErrorCode: 400, Description: "Bad Request: file is too big"},
	"invalid_file_id":         {ErrorCode: 400, Description: "Bad Request: invalid file id"},
	"member_not_found":        {ErrorCode: 400, Description: "Bad Request: member not found"},
	"group_deactivated":       {ErrorCode: 400, Description: "Bad Request: group is deactivated"},
	"peer_id_invalid":         {ErrorCode: 400, Description: "Bad Request: PEER_ID_INVALID"},
	"wrong_parameter_action":  {ErrorCode: 400, Description: "Bad Request: wrong parameter action in request"},

	// 401 Unauthorized
	"unauthorized": {ErrorCode: 401, Description: "Unauthorized"},

	// 403 Forbidden
	"forbidden":               {ErrorCode: 403, Description: "Forbidden"},
	"bot_blocked":             {ErrorCode: 403, Description: "Forbidden: bot was blocked by the user"},
	"bot_kicked":              {ErrorCode: 403, Description: "Forbidden: bot was kicked from the chat"},
	"cant_initiate":           {ErrorCode: 403, Description: "Forbidden: bot can't initiate conversation with a user"},
	"cant_send_to_bots":       {ErrorCode: 403, Description: "Forbidden: bot can't send messages to bots"},
	"not_member_channel":      {ErrorCode: 403, Description: "Forbidden: bot is not a member of the channel chat"},
	"not_member_supergroup":   {ErrorCode: 403, Description: "Forbidden: bot is not a member of the supergroup chat"},
	"user_deactivated":        {ErrorCode: 403, Description: "Forbidden: user is deactivated"},
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
