package faker

import (
	"fmt"
	"strings"
	"time"
)

// generateString generates a realistic string value based on the field name.
func (f *Faker) generateString(fieldName string) string {
	name := strings.ToLower(fieldName)

	// ID-like fields
	if strings.HasSuffix(name, "_id") || strings.HasSuffix(name, "id") {
		return f.generateFileID()
	}

	// Username fields
	if name == "username" {
		return f.generateUsername()
	}

	// Name fields
	if name == "first_name" {
		return f.RandomChoice(firstNames)
	}
	if name == "last_name" {
		return f.RandomChoice(lastNames)
	}
	if name == "name" || name == "title" {
		return f.generateTitle()
	}

	// Text content
	if name == "text" || name == "caption" || name == "description" {
		return f.generateText()
	}

	// URLs
	if strings.HasSuffix(name, "_url") || name == "url" || name == "href" {
		return f.generateURL()
	}

	// Paths
	if strings.HasSuffix(name, "_path") || name == "file_path" {
		return f.generateFilePath()
	}

	// Email
	if name == "email" {
		return f.generateEmail()
	}

	// Phone
	if name == "phone_number" {
		return f.generatePhoneNumber()
	}

	// Language
	if name == "language_code" {
		return f.RandomChoice(languageCodes)
	}

	// Type/status fields
	if name == "type" || name == "status" {
		return "default"
	}

	// Currency
	if name == "currency" {
		return f.RandomChoice(currencies)
	}

	// Emoji
	if name == "emoji" || strings.HasSuffix(name, "_emoji") {
		return f.RandomChoice(emojis)
	}

	// Command
	if name == "command" {
		return f.generateCommand()
	}

	// Query
	if name == "query" || name == "inline_query" {
		return f.generateQuery()
	}

	// MIME type
	if name == "mime_type" {
		return f.RandomChoice(mimeTypes)
	}

	// Performer/Author
	if name == "performer" || name == "author" || name == "author_signature" {
		return f.generateAuthor()
	}

	// Default: use field name as base
	return fmt.Sprintf("%s_%d", fieldName, f.rng.Intn(10000))
}

// generateInt64 generates a realistic int64 value based on the field name.
func (f *Faker) generateInt64(fieldName string) int64 {
	name := strings.ToLower(fieldName)

	// ID fields - large numbers
	if strings.HasSuffix(name, "_id") || name == "id" {
		return f.RandomInt64(100000000, 999999999)
	}

	// Date/time fields - Unix timestamps
	if name == "date" || strings.HasSuffix(name, "_date") || strings.HasSuffix(name, "_time") {
		return time.Now().Unix()
	}

	// Count fields - small numbers
	if strings.HasSuffix(name, "_count") || name == "count" || name == "total_count" {
		return f.RandomInt64(1, 100)
	}

	// Size fields
	if strings.HasSuffix(name, "_size") || name == "size" || name == "file_size" {
		return f.RandomInt64(1024, 1024*1024)
	}

	// Duration
	if name == "duration" {
		return f.RandomInt64(1, 600)
	}

	// Dimensions
	if name == "width" || name == "height" {
		return f.RandomInt64(100, 1920)
	}

	// Offset/position
	if name == "offset" || name == "length" {
		return f.RandomInt64(0, 100)
	}

	// Message thread
	if name == "message_thread_id" {
		return f.RandomInt64(1, 10000)
	}

	// Default: small positive number
	return f.RandomInt64(1, 1000)
}

// generateFloat64 generates a realistic float64 value based on the field name.
func (f *Faker) generateFloat64(fieldName string) float64 {
	name := strings.ToLower(fieldName)

	// Coordinates
	if name == "latitude" {
		return f.RandomFloat64(-90.0, 90.0)
	}
	if name == "longitude" {
		return f.RandomFloat64(-180.0, 180.0)
	}

	// Heading/direction
	if name == "heading" {
		return f.RandomFloat64(0, 360)
	}

	// Accuracy
	if name == "horizontal_accuracy" || strings.HasSuffix(name, "_accuracy") {
		return f.RandomFloat64(0, 100)
	}

	// Score/rating
	if name == "score" || name == "rating" {
		return f.RandomFloat64(0, 5)
	}

	// Scale
	if name == "scale" || name == "zoom" {
		return f.RandomFloat64(1, 20)
	}

	// Default: small decimal
	return f.RandomFloat64(0, 100)
}

// generateBool generates a boolean value based on the field name.
func (f *Faker) generateBool(fieldName string) bool {
	name := strings.ToLower(fieldName)

	// is_bot is always true for bot responses
	if name == "is_bot" {
		return true
	}

	// Common "is_" prefixes that are often true
	trueBiased := []string{
		"is_premium", "can_join_groups", "can_read_all_group_messages",
		"supports_inline_queries", "is_anonymous", "can_manage_chat",
	}
	for _, prefix := range trueBiased {
		if name == prefix {
			return f.RandomBool(0.7) // 70% chance of true
		}
	}

	// Common "is_" prefixes that are often false
	falseBiased := []string{
		"is_closed", "is_banned", "is_restricted", "is_deleted",
		"is_revoked", "is_member",
	}
	for _, prefix := range falseBiased {
		if name == prefix {
			return f.RandomBool(0.3) // 30% chance of true
		}
	}

	// Default: 50/50
	return f.RandomBool(0.5)
}

// Helper generators

func (f *Faker) generateFileID() string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"
	length := 40 + f.rng.Intn(20)
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[f.rng.Intn(len(chars))]
	}
	return "AgACAgIAAxk" + string(result)
}

func (f *Faker) generateUsername() string {
	adjective := f.RandomChoice(usernameAdjectives)
	noun := f.RandomChoice(usernameNouns)
	num := f.rng.Intn(1000)
	return fmt.Sprintf("%s_%s_%d", adjective, noun, num)
}

func (f *Faker) generateTitle() string {
	adjective := f.RandomChoice(titleAdjectives)
	noun := f.RandomChoice(titleNouns)
	return fmt.Sprintf("%s %s", adjective, noun)
}

func (f *Faker) generateText() string {
	sentences := 1 + f.rng.Intn(3)
	var parts []string
	for i := 0; i < sentences; i++ {
		parts = append(parts, f.RandomChoice(sampleSentences))
	}
	return strings.Join(parts, " ")
}

func (f *Faker) generateURL() string {
	domain := f.RandomChoice(domains)
	path := f.RandomChoice(urlPaths)
	return fmt.Sprintf("https://%s/%s", domain, path)
}

func (f *Faker) generateFilePath() string {
	folder := f.RandomChoice(folders)
	ext := f.RandomChoice(extensions)
	name := fmt.Sprintf("file_%d", f.rng.Intn(10000))
	return fmt.Sprintf("%s/%s.%s", folder, name, ext)
}

func (f *Faker) generateEmail() string {
	name := strings.ToLower(f.RandomChoice(firstNames))
	domain := f.RandomChoice(emailDomains)
	num := f.rng.Intn(100)
	return fmt.Sprintf("%s%d@%s", name, num, domain)
}

func (f *Faker) generatePhoneNumber() string {
	countryCode := f.RandomChoice(countryCodes)
	number := fmt.Sprintf("%d%d%d%d%d%d%d%d%d%d",
		f.rng.Intn(10), f.rng.Intn(10), f.rng.Intn(10),
		f.rng.Intn(10), f.rng.Intn(10), f.rng.Intn(10),
		f.rng.Intn(10), f.rng.Intn(10), f.rng.Intn(10),
		f.rng.Intn(10))
	return fmt.Sprintf("+%s%s", countryCode, number)
}

func (f *Faker) generateCommand() string {
	return "/" + f.RandomChoice(commands)
}

func (f *Faker) generateQuery() string {
	return f.RandomChoice(queries)
}

func (f *Faker) generateAuthor() string {
	first := f.RandomChoice(firstNames)
	last := f.RandomChoice(lastNames)
	return fmt.Sprintf("%s %s", first, last)
}

// Data sets for generation

var firstNames = []string{
	"Alex", "Emma", "James", "Sophia", "Michael", "Olivia", "William", "Ava",
	"John", "Isabella", "David", "Mia", "Richard", "Charlotte", "Joseph", "Amelia",
	"Thomas", "Harper", "Daniel", "Evelyn", "Matthew", "Abigail", "Andrew", "Emily",
}

var lastNames = []string{
	"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
	"Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson",
	"Thomas", "Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson",
}

var usernameAdjectives = []string{
	"cool", "fast", "smart", "happy", "lucky", "super", "mega", "ultra",
	"pro", "elite", "top", "best", "great", "awesome", "epic", "legend",
}

var usernameNouns = []string{
	"user", "coder", "dev", "hacker", "ninja", "wizard", "master", "guru",
	"bot", "player", "gamer", "creator", "maker", "builder", "runner", "rider",
}

var titleAdjectives = []string{
	"Official", "Amazing", "Awesome", "Great", "Super", "Best", "Top", "Premium",
	"Elite", "Pro", "Ultimate", "Fantastic", "Incredible", "Wonderful", "Excellent",
}

var titleNouns = []string{
	"Group", "Channel", "Community", "Team", "Club", "Network", "Hub", "Center",
	"Zone", "Space", "Place", "World", "Universe", "Kingdom", "Empire",
}

var sampleSentences = []string{
	"Hello, this is a test message.",
	"Welcome to the group!",
	"Thank you for your message.",
	"That's a great question!",
	"I'll get back to you soon.",
	"Please check the documentation.",
	"Have a wonderful day!",
	"Let me know if you need help.",
}

var domains = []string{
	"example.com", "test.org", "sample.net", "demo.io", "mock.dev",
	"telegram.org", "api.example.com", "cdn.example.com",
}

var urlPaths = []string{
	"page", "article/123", "post/456", "image.jpg", "document.pdf",
	"api/v1/data", "files/download", "media/photo",
}

var folders = []string{
	"photos", "documents", "videos", "voice", "stickers", "animations", "music",
}

var extensions = []string{
	"jpg", "png", "gif", "mp4", "mp3", "ogg", "pdf", "doc", "webp", "tgs",
}

var emailDomains = []string{
	"example.com", "test.org", "mail.com", "email.net", "inbox.io",
}

var countryCodes = []string{
	"1", "7", "44", "49", "33", "39", "34", "81", "86", "91",
}

var languageCodes = []string{
	"en", "ru", "de", "fr", "es", "it", "pt", "ja", "zh", "ko", "ar", "hi",
}

var currencies = []string{
	"USD", "EUR", "GBP", "RUB", "JPY", "CNY", "INR", "BRL", "KRW", "TRY",
}

var emojis = []string{
	"👍", "❤️", "🔥", "👏", "😊", "🎉", "💪", "✨", "🚀", "💯",
}

var mimeTypes = []string{
	"image/jpeg", "image/png", "image/gif", "video/mp4", "audio/mpeg",
	"audio/ogg", "application/pdf", "text/plain", "application/json",
}

var commands = []string{
	"start", "help", "settings", "about", "cancel", "menu", "status", "info",
}

var queries = []string{
	"search query", "example", "test", "hello world", "sample text",
}
