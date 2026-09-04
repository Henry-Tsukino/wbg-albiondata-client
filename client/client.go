package client

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"
)

var version string
var startupMessages []string
var autoStartStatus string = "Disabled"
var statusActive bool = false
var IPinok string = "https://hermes-production-817b.up.railway.app"

// ANSI color codes
const (
	colorYellow = "\033[38;5;220m" // Желтый цвет близкий к #ffb12c
	colorCyan   = "\033[36m"       // Cyan
	colorGreen  = "\033[32m"       // Green
	colorWhite  = "\033[37m"       // White
	colorReset  = "\033[0m"        // Reset
)

// Client struct base
type Client struct {
}

// NewClient return a new Client instance
func NewClient(_version string) *Client {
	version = _version
	return &Client{}
}

// colorify adds color to a string with an optional colored square
func colorify(color, text, square string) string {
	if square != "" {
		return fmt.Sprintf("%s%s%s%s", color, square, text, colorReset)
	}
	return fmt.Sprintf("%s%s%s", color, text, colorReset)
}

// getRandomJoke returns a random joke
func getRandomJoke() string {
	jokes := []string{
		"Муж спрашивает жену: 'Ты любишь меня?' Жена: 'Конечно!' Муж: 'А почему мы никогда не говорим о наших чувствах?' Жена: 'Потому что у меня нет времени на фантазии!'",
		"Почему женщины не могут играть в футбол? Потому что они уходят в крайний случай!",
		"Учитель спрашивает: 'Почему ты не выучил уроки?' Ученик: 'Интернет был медленный!' Учитель: 'А и до интернета были ленивые дети!'",
		"Мужчина в парикмахерской: 'Покороче!' Парикмахер: 'На сколько?' Мужчина: 'Пока жена не заметит!'",
		"Жена: 'Ты когда-нибудь меня любил?' Муж: 'Конечно!' Жена: 'А когда это было?' Муж: 'До того, как мы поженились!'",
		"Почему программисты не выходят из дома? Потому что <home> - это корневая директория!",
		"Папа сыну: 'В твои годы я уже в школу ходил!' Сын: 'Ну и где он сейчас, твой класс?'",
		"Женщина говорит мужу: 'Я задумана!' Муж: 'Это временно или навсегда?'",
		"Как называется гирлянда, которая не светит? Провод! А как гирлянда, которая светит? Проводка!",
		"Начальник спрашивает работника: 'Почему ты опаздываешь?' Работник: 'Будильник был неправ!'",
		"Жена мужу: 'Думаешь, я красивая?' Муж: 'Конечно!' Жена: 'А без макияжа?' Муж: 'I don't know - я никогда это не видел!'",
		"Подходит программист к девушке: 'Можно я буду 'if' в твоей жизни?' Девушка: 'Нет, ты просто 'else'!'",
		"Врач пациенту: 'У вас в крови инфекция!' Пациент: 'Не может быть! Я же никогда его не видел!'",
		"Почему у психолога нет друзей? Потому что все его слова - 'и это нормально!'",
		"Муж приходит домой: 'Я нашел работу снов!' Жена: 'Какую?' Муж: 'Ночной сторож!'",
		"Минус социальных сетей - это то, что каждый может что-то про тебя сказать. Плюс - что мало кто это читает!",
		"Сын спрашивает отца: 'Папа, ты богат?' Отец: 'Нет, сын.' Сын: 'А почему ты не попросишь денег в маме?'",
		"Как назвать кота без хвоста? Мяукаратель!",
	}

	return jokes[rand.Intn(len(jokes))]
}

// printBanner outputs a beautiful ASCII art banner with system information
func printBanner() {
	banner := []string{
		"                       .7!ь.                                    ",
		"                     .?BBBG7ь.                                  ",
		"                    .JBBBBBBBB?.                                ",
		"                  :YBBBBBBBBBBBBJ.                              ",
		"                :YBBBBBBBBBBBBBBBBY:                            ",
		"             .^5BBBBBBBBBBBBBBBBBBBB5^.                         ",
		"           .  !GBBBBBBBBBBBBBBBBBBBBB7                          ",
		"         .JBP~  ~PBBBBBBBBBBBBBBBBG!  ~PB?.                     ",
		"       :YBBBBBG!  :5BBBBBBBBBBBBP~  ~GBBBBBJ.                   ",
		"     :YBBBBBBBBBG7  :YBBBBBBBBP^  !GBBBBBBBBBY:                 ",
		"   ^5BBBBBBBBBBBBBB?. .JBBBB5:  7GBBBBBBBBBBBBB5^               ",
		" ^PBBBBBBBBBBBBBBBBBBY. .?Y: .?BBBBBBBBBBBBBBBBBBP^             ",
		"5#BBBBBBBBBBBBBBBBBBB##J    !B#BBBBBBBBBBBBBBBBBBB#Y            ",
		" !PBBBBBBBBBBBBBBBBBBP^  77  :5BBBBBBBBBBBBBBBBBBP~             ",
		"   ^PBBBBBBBBBBBBBB5:  7BBBB?. .YBBBBBBBBBBBBBB5^               ",
		"     ^5BBBBBBBBBBY: .?BBBBBBBBJ. .JBBBBBBBBBB5:                 ",
		"       :YBBBBBBJ. .JBBBBBBBBBBBBY: .?BBBBBBY:                   ",
		"         .?BB?. :YBBBBBBBBBBBBBBBB5:  7GBJ.                     ",
		"           .. ^5BBBBBBBBBBBBBBBBBBBBP~  .                       ",
		"             .YBBBBBBBBBBBBBBBBBBBBBBY.                         ",
		"               .?BBBBBBBBBBBBBBBBBB?.                           ",
		"                  7GBBBBBBBBBBBBG7.                             ",
		"                    ~PBBBBBBBBG!                                ",
		"                      ^5BBBBP~                                  ",
		"                        :Y5^                                    ",
	}

	currentTime := time.Now().Format("15:04:05")
	joke := getRandomJoke()

	// Build info lines - combine header with status
	info := []string{
		"",
		colorify(colorCyan, "  ■ WBG-AODP Client v"+version, ""),
		colorify(colorWhite, "  ■ Time: "+currentTime, ""),
		colorify(colorCyan, "  ■ OS: "+runtime.GOOS, ""),
		colorify(colorGreen, "  ■ Watching Albion", ""),
	}

	// Add status if player is active
	if statusActive {
		info = append(info, colorify(colorGreen, "  ■ Status: active", ""))
	}

	// Add auto-start status with color based on status
	autoStartColor := colorYellow
	if autoStartStatus == "Enabled" {
		autoStartColor = colorGreen
	}
	info = append(info, colorify(autoStartColor, "  ■ Auto-start: "+autoStartStatus, ""))
	info = append(info, "")

	// Add startup messages (up to 6 lines)
	for i := 0; i < len(startupMessages) && i < 6; i++ {
		info = append(info, colorify(colorWhite, "  ■ "+startupMessages[i], ""))
	}

	maxLines := len(banner)
	if len(info) > maxLines {
		maxLines = len(info)
	}

	// Print combined output with colored ASCII art
	for i := 0; i < maxLines; i++ {
		left := ""
		if i < len(banner) {
			left = colorYellow + banner[i] + colorReset
		}

		right := ""
		if i < len(info) {
			right = info[i]
		}

		fmt.Println(left + right)
	}

	// Print joke with word wrapping if needed
	fmt.Println("")
	fmt.Printf("  💡 "+colorYellow+"%s"+colorReset+"\n", joke)
	fmt.Println("")
}

// repaintBanner clears the screen and redraws the banner
func repaintBanner() {
	fmt.Print("\033[2J\033[H")
	printBanner()
}

// UpdateAutoStartStatus updates the auto-start status display
func UpdateAutoStartStatus(status string) {
	autoStartStatus = status
	repaintBanner()
}

// SetPlayerActive marks player as active
func SetPlayerActive() {
	statusActive = true
	repaintBanner()
}

// Run starts client settings and run
func (client *Client) Run() error {
	// Collect startup messages to display with banner
	startupMessages = []string{
		"Third-party application - no affiliation",
		"Parameters: use -h for help",
		"Windows: network adapter retry enabled",
	}

	printBanner()

	ConfigGlobal.setupDebugEvents()
	ConfigGlobal.setupDebugOperations()

	createDispatcher()

	if ConfigGlobal.Offline {
		processOffline(ConfigGlobal.OfflinePath)

		// Allow time for any async uploads/processing to complete, then exit.
		time.Sleep(10 * time.Second)
		os.Exit(0)

	} else {
		apw := newAlbionProcessWatcher()
		return apw.run()
	}
	return nil
}
