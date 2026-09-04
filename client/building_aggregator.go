package client

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ao-data/albiondata-client/log"
)

// -----------------------------------------------------------------------
// Ивенты, которые агрегируем (номера и раскладка полей взяты из debug-лога):
//
//   45  evNewBuilding       -> type(3), owner(11,12)
//   49  evCraftBuildingInfo -> price(1) - текущая публичная цена, friendPrice(2) - старая
//   210 evAccessStatus      -> roles(5) сопоставляются с values(6), берём запись "friend"
//
// Важно: evCraftBuildingInfo сервер шлёт не пассивно (просто когда станок
// в зоне видимости), а только когда игрок реально открывает у него
// интерфейс крафта/цены. Это почти всегда происходит позже, чем
// evNewBuilding + buildingFlushDelay, поэтому цена приходит "запоздало" -
// см. sent/lastSent*-поля в BuildingEntity и логику повторной отправки в
// merge()/flush() ниже.
//
// Ключ сущности - id (params[0]), общий для всех трёх событий.
// -----------------------------------------------------------------------

const (
	evtNewBuilding       = byte(45)
	evtCraftBuildingInfo = byte(49)
	evtAccessStatus      = byte(210)

	buildingStaleAfter = 10 * time.Second        // шаг 7: чистка "зависших", если база (type/owner) так и не пришла
	buildingFlushDelay = 1000 * time.Millisecond // шаг 5: защитный таймер - сколько ждать доп.полей после базы

	// buildingResendDelay - задержка перед повторной отправкой, когда
	// станок уже был отправлен один раз, но пришло новое значимое
	// обновление (чаще всего - evCraftBuildingInfo, который сервер шлёт
	// только когда игрок реально открывает интерфейс станка, а это почти
	// всегда происходит позже, чем buildingFlushDelay после evNewBuilding).
	buildingResendDelay = 300 * time.Millisecond

	// buildingIdleAfterSent - сколько держать в памяти уже отправленный
	// станок в ожидании возможных поздних обновлений (цена/друзья), прежде
	// чем окончательно забыть о нём. Значительно больше buildingStaleAfter,
	// потому что "база" тут уже есть, риск утечки памяти невелик, а цена
	// может прийти спустя минуты после первого появления станка.
	buildingIdleAfterSent = 5 * time.Minute
)

// BuildingEntity - агрегированная сущность станка, собранная из нескольких ивентов.
type BuildingEntity struct {
	ID        int32
	Type      string
	OwnerID   string
	OwnerName string

	X int32 // координата X станка на карте (evNewBuilding, params[4][0])
	Y int32 // координата Y станка на карте (evNewBuilding, params[4][1])

	Price       int64 // текущая публичная цена (evCraftBuildingInfo, params[1])
	friendPrice int64 // старая цена (evCraftBuildingInfo, params[2])

	Friend string // значение params[6], у которого params[5] == "friend"

	lastUpdate time.Time
	haveBase   bool // type и owner уже пришли (т.е. было evNewBuilding)
	flushTimer *time.Timer

	// sent - сущность уже была отправлена хотя бы раз. После этого мы её не
	// удаляем из pending, а держим дальше: evCraftBuildingInfo (цена) почти
	// всегда приходит позже первого флаша, и если её не с чем будет
	// смёржить, она потеряется.
	sent bool
	// lastSentPrice/lastSentfriendPrice/lastSentFriend - снимок того, что уже
	// было отправлено, чтобы не устраивать повторную отправку на каждое
	// пустое обновление (например, повторный evAccessStatus без новых
	// данных), а только когда реально появилось что-то новое.
	lastSentPrice       int64
	lastSentfriendPrice int64
	lastSentFriend      string
}

func newBuildingEntity(id int32) *BuildingEntity {
	return &BuildingEntity{
		ID:         id,
		lastUpdate: time.Now(),
	}
}

// -----------------------------------------------------------------------
// Шаг 3: по одному парсеру на каждый evType.
// Никакой логики агрегации внутри - только "сырые поля -> именованный объект".
// -----------------------------------------------------------------------

func parseNewBuilding(params map[byte]interface{}) (int32, *BuildingEntity, error) {
	id, ok := toInt32Field(params[0])
	if !ok {
		return 0, nil, fmt.Errorf("evNewBuilding: нет корректного id в params[0]")
	}

	upd := newBuildingEntity(id)

	if typ, ok := params[3].(string); ok {
		upd.Type = typ
	}
	if x, y, ok := toCoordsField(params[4]); ok {
		upd.X = x
		upd.Y = y
	}
	if owner, ok := params[11].(string); ok {
		upd.OwnerID = owner
	}
	if name, ok := params[12].(string); ok {
		upd.OwnerName = name
	}

	upd.haveBase = upd.Type != "" && (upd.OwnerID != "" || upd.OwnerName != "")

	return id, upd, nil
}

func parseCraftBuildingInfo(params map[byte]interface{}) (int32, *BuildingEntity, error) {
	id, ok := toInt32Field(params[0])
	if !ok {
		return 0, nil, fmt.Errorf("evCraftBuildingInfo: нет корректного id в params[0]")
	}

	upd := newBuildingEntity(id)

	if price, ok := toInt64Field(params[1]); ok {
		upd.Price = price
	}
	if old, ok := toInt64Field(params[2]); ok {
		upd.friendPrice = old
	}

	return id, upd, nil
}

func parseAccessStatus(params map[byte]interface{}) (int32, *BuildingEntity, error) {
	id, ok := toInt32Field(params[0])
	if !ok {
		return 0, nil, fmt.Errorf("evAccessStatus: нет корректного id в params[0]")
	}

	upd := newBuildingEntity(id)

	roles, rok := params[5].([]string)
	values, vok := params[6].([]string)
	if rok && vok && len(roles) == len(values) {
		for i, r := range roles {
			if r == "friend" {
				upd.Friend = values[i]
			}
		}
	}

	return id, upd, nil
}

// -----------------------------------------------------------------------
// Хелперы разбора отдельных полей (photon params приходят как interface{}
// разных конкретных типов в зависимости от кодека)
// -----------------------------------------------------------------------

// toInt32Field и toInt64Field используют reflect, а не список конкретных типов.
//
// Причина: изначально здесь был switch по конкретным типам (int32, int,
// int16, uint16, float64), и он пропускал маленькие значения вроде id=20 -
// Photon-парсер отдаёт такие числа как byte/uint8 (а иногда int8/uint32/
// uint64/uint в зависимости от кодека), и ни один из этих типов в списке не
// было. Из-за этого evNewBuilding с params[0]=20 (byte) проваливался с
// "нет корректного id в params[0]", хотя данные были корректны. Через
// reflect.Kind покрываем весь диапазон целочисленных типов сразу, без
// необходимости угадывать, какой именно тип отдаёт кодек в конкретном случае.
func toInt32Field(v interface{}) (int32, bool) {
	n, ok := toInt64Field(v)
	if !ok {
		return 0, false
	}
	return int32(n), true
}

func toInt64Field(v interface{}) (int64, bool) {
	if v == nil {
		return 0, false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return int64(rv.Float()), true
	}
	return 0, false
}

// toCoordsField разбирает пару координат вида [185 65] - Photon отдаёт такие
// поля как срез (slice) из двух чисел, но конкретный тип элементов
// (int16/float64/... - зависит от кодека) заранее не известен, поэтому,
// как и в toInt64Field, читаем через reflect вместо жёсткого списка типов.
func toCoordsField(v interface{}) (x, y int32, ok bool) {
	if v == nil {
		return 0, 0, false
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return 0, 0, false
	}
	if rv.Len() != 2 {
		return 0, 0, false
	}
	xi, xok := toInt64Field(rv.Index(0).Interface())
	yi, yok := toInt64Field(rv.Index(1).Interface())
	if !xok || !yok {
		return 0, 0, false
	}
	return int32(xi), int32(yi), true
}

// -----------------------------------------------------------------------
// Шаг 4-7: хранилище незавершённых сущностей, merge, флаш готовых,
// чистка зависших.
// -----------------------------------------------------------------------

// BuildingReadyFunc вызывается, когда сущность отправлена (по таймеру).
type BuildingReadyFunc func(*BuildingEntity)

type buildingAggregator struct {
	mu      sync.Mutex
	pending map[int32]*BuildingEntity

	onReady BuildingReadyFunc

	// Отправка идёт через общий uploader interface (см. uploader.go/
	// uploader_http.go) вместо самодельного http.Post - тот же путь, что и у
	// остального клиента (sendToIngest), с общей логикой транспорта/keep-alive.
	uploader uploader

	quit chan struct{}

	// счётчики для периодической сводки (см. statsLoop) - только для того,
	// чтобы можно было глазами понять "работает / не работает" без debug-логов.
	// ВАЖНО: sendToIngest ничего не возвращает (fire-and-forget, ошибки только
	// логируются внутри uploader'а), поэтому sentErr здесь считает только
	// локальные ошибки маршалинга JSON - настоящие ошибки самой отправки надо
	// смотреть в логах uploader'а ("Error while sending ingest...").
	eventsReceived int64
	sentOK         int64
	sentErr        int64
}

func newBuildingAggregator(onReady BuildingReadyFunc, up uploader) *buildingAggregator {
	a := &buildingAggregator{
		pending:  make(map[int32]*BuildingEntity),
		onReady:  onReady,
		uploader: up,
		quit:     make(chan struct{}),
	}
	log.Debugf("BuildingAggregator: запущен, готовые станки будут отправляться через ingest на %s", IPinok+"/prices")
	go a.cleanupLoop()
	go a.statsLoop()
	return a
}

// statsLoop раз в 30 секунд печатает сводку: сколько сырых событий (45/49/210)
// пришло, сколько станков реально ушло боту и сколько было ошибок отправки.
// Если тут постоянно "получено=0" - значит события вообще не долетают до
// merge() (см. dispatchBuildingEvent в listener.go). Если "отправлено=0" при
// ненулевом "получено" - станки не набирают базу (type+owner) или бот не
// отвечает (см. отдельные priceBot: ... строки).
func (a *buildingAggregator) statsLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-a.quit:
			return
		case <-ticker.C:
			received := atomic.SwapInt64(&a.eventsReceived, 0)
			ok := atomic.SwapInt64(&a.sentOK, 0)
			errs := atomic.SwapInt64(&a.sentErr, 0)
			a.mu.Lock()
			pendingCount := len(a.pending)
			a.mu.Unlock()
			log.Debugf("BuildingAggregator: за 30с получено=%d, отправлено=%d, ошибок отправки=%d, сейчас в ожидании=%d",
				received, ok, errs, pendingCount)
		}
	}
}

// merge сливает частичное обновление upd (для сущности id) в накопленное
// состояние.
//
//   - Первое появление базы (type+owner, т.е. evNewBuilding) запускает
//     разовый таймер buildingFlushDelay - он даёт цене (49) и friend-доступу
//     (210) время подтянуться, после чего сущность уходит в onReady.
//   - Сущность НЕ удаляется после первой отправки (см. flush) - если позже
//     придёт значимое обновление (обычно цена - evCraftBuildingInfo почти
//     всегда приходит позже первого флаша, когда игрок открывает интерфейс
//     станка), запускается короткий таймер buildingResendDelay, и повторно
//     отправляется уже полная сущность с актуальной ценой.
func (a *buildingAggregator) merge(id int32, upd *BuildingEntity) {
	atomic.AddInt64(&a.eventsReceived, 1)

	a.mu.Lock()
	defer a.mu.Unlock()

	existing, ok := a.pending[id]
	if !ok {
		log.Debugf("BuildingAggregator: получено новое событие по станку id=%d", id)
		existing = newBuildingEntity(id)
		a.pending[id] = existing
	}

	if upd.Type != "" {
		existing.Type = upd.Type
	}
	if upd.OwnerID != "" {
		existing.OwnerID = upd.OwnerID
	}
	if upd.OwnerName != "" {
		existing.OwnerName = upd.OwnerName
	}
	if upd.X != 0 || upd.Y != 0 {
		existing.X = upd.X
		existing.Y = upd.Y
	}
	if upd.Price != 0 {
		existing.Price = upd.Price
	}
	if upd.friendPrice != 0 {
		existing.friendPrice = upd.friendPrice
	}
	if upd.Friend != "" {
		existing.Friend = upd.Friend
	}
	if upd.haveBase {
		existing.haveBase = true
	}
	existing.lastUpdate = time.Now()

	log.Debugf("BuildingAggregator: смёржено в id=%d, type=%s, xy=[%d %d], price=%d, friendPrice=%d, friend=%q",
		id, existing.Type, existing.X, existing.Y, existing.Price, existing.friendPrice, existing.Friend)

	if !existing.haveBase || existing.flushTimer != nil {
		return
	}

	if !existing.sent {
		log.Debugf("BuildingAggregator: id=%d набрал базу (type=%s, owner=%s), жду %v перед отправкой",
			id, existing.Type, existing.OwnerName, buildingFlushDelay)
		existing.flushTimer = time.AfterFunc(buildingFlushDelay, func() {
			a.flush(id)
		})
		return
	}

	// Уже отправляли раньше - шлём повтор только если появилось что-то
	// новое по сравнению с тем, что уже ушло боту.
	if existing.Price != existing.lastSentPrice ||
		existing.friendPrice != existing.lastSentfriendPrice ||
		existing.Friend != existing.lastSentFriend {
		log.Debugf("BuildingAggregator: id=%d - позднее обновление (price=%d->%d, friendPrice=%d->%d, friend=%q->%q), пересылаю через %v",
			id, existing.lastSentPrice, existing.Price, existing.lastSentfriendPrice, existing.friendPrice,
			existing.lastSentFriend, existing.Friend, buildingResendDelay)
		existing.flushTimer = time.AfterFunc(buildingResendDelay, func() {
			a.flush(id)
		})
	}
}

// flush отправляет текущее состояние сущности в onReady - вызывается по
// таймеру из merge() (как при первой отправке, так и при повторной, если
// после неё пришло что-то новое).
//
// Сущность больше НЕ удаляется из pending при первой отправке (в отличие от
// прежней версии) - она остаётся там, чтобы поздние обновления (обычно
// evCraftBuildingInfo с ценой, которое сервер шлёт только когда игрок
// реально открывает станок - т.е. почти всегда позже buildingFlushDelay)
// могли смёржиться и уйти повторным флашем, а не потеряться. Забывает о
// станке только cleanupLoop, и то с гораздо большим таймаутом после первой
// отправки (buildingIdleAfterSent).
func (a *buildingAggregator) flush(id int32) {
	a.mu.Lock()
	ent, ok := a.pending[id]
	if !ok {
		a.mu.Unlock()
		return
	}
	ent.flushTimer = nil
	ent.sent = true
	ent.lastSentPrice = ent.Price
	ent.lastSentfriendPrice = ent.friendPrice
	ent.lastSentFriend = ent.Friend
	snapshot := *ent
	a.mu.Unlock()

	log.Debugf("BuildingAggregator: id=%d готов, отправляю (type=%s, owner=%s, xy=[%d %d], price=%d, friendPrice=%d, friend=%q)",
		id, snapshot.Type, snapshot.OwnerName, snapshot.X, snapshot.Y, snapshot.Price, snapshot.friendPrice, snapshot.Friend)

	if a.onReady != nil {
		a.onReady(&snapshot)
	}
}

// cleanupLoop - шаг 7: раз в 2 секунды выкидываем "зависшие" записи из
// pending. Два случая:
//   - база (type/owner) так и не появилась дольше buildingStaleAfter -
//     как и раньше, это мусор, который никогда не отправится.
//   - сущность уже была отправлена (sent), но с последнего обновления
//     прошло дольше buildingIdleAfterSent - станок, видимо, давно вне
//     зоны видимости, и ждать по нему запоздалую цену дальше смысла нет.
func (a *buildingAggregator) cleanupLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-a.quit:
			return
		case <-ticker.C:
			a.mu.Lock()
			for id, ent := range a.pending {
				idle := time.Since(ent.lastUpdate)
				if !ent.haveBase && idle > buildingStaleAfter {
					log.Debugf("BuildingAggregator: удаляю зависший id=%d (базы так и не было)", id)
					delete(a.pending, id)
					continue
				}
				if ent.sent && idle > buildingIdleAfterSent {
					log.Debugf("BuildingAggregator: удаляю id=%d из памяти (уже отправлен, %v без обновлений)", id, idle)
					delete(a.pending, id)
				}
			}
			a.mu.Unlock()
		}
	}
}

func (a *buildingAggregator) stop() {
	close(a.quit)
}

// -----------------------------------------------------------------------
// Шаг 6: отправка готовой сущности через общий ingest-путь (uploader),
// на personal price-bot (см. IPinok в client.go).
// -----------------------------------------------------------------------

type buildingUpload struct {
	AlbionId    int32  `json:"AlbionId"`
	Type        string `json:"Type"`
	OwnerId     string `json:"OwnerId"`
	OwnerName   string `json:"OwnerName"`
	X           int32  `json:"X"`
	Y           int32  `json:"Y"`
	Price       int64  `json:"Price"`
	friendPrice int64  `json:"friendPrice"`
	Friend      string `json:"Friend,omitempty"`
	LocationId  string `json:"LocationId"` // текущая локация (зона) игрока на момент отправки
}

func (a *buildingAggregator) sendBuildingUpdate(state *albionState, b *BuildingEntity) {
	upload := buildingUpload{
		AlbionId:    b.ID,
		Type:        b.Type,
		OwnerId:     b.OwnerID,
		OwnerName:   b.OwnerName,
		X:           b.X,
		Y:           b.Y,
		Price:       b.Price,
		friendPrice: b.friendPrice,
		Friend:      b.Friend,
		LocationId:  state.LocationId,
	}

	// Это данные под персональный price-bot, а не под официальный протокол
	// albion-data - поэтому шлём только туда, не в sendMsgToPublicUploaders.
	go a.sendViaIngest(state, upload)
}

// sendViaIngest сериализует станок и отдаёт его общему uploader'у
// (httpUploader.sendToIngest), тем же путём, каким остальной клиент шлёт
// данные в ingest - топик "prices" даёт тот же итоговый URL, что и раньше
// (IPinok + "/prices", см. newHTTPUploader(IPinok) в router.go).
func (a *buildingAggregator) sendViaIngest(state *albionState, upload buildingUpload) {
	data, err := json.Marshal(upload)
	if err != nil {
		atomic.AddInt64(&a.sentErr, 1)
		log.Errorf("priceBot: ошибка маршалинга станка id=%v: %v", upload.AlbionId, err)
		return
	}

	if a.uploader == nil {
		atomic.AddInt64(&a.sentErr, 1)
		log.Errorf("priceBot: uploader не настроен, станок id=%v не отправлен", upload.AlbionId)
		return
	}

	log.Debugf("priceBot: отправляю станок id=%v через ingest (topic=prices): %s", upload.AlbionId, string(data))

	// identifier используется только для http+pow вариантов uploader'а
	// (см. комментарий в uploader_http.go); для нашего personal price-bot
	// достаточно ID станка.
	a.uploader.sendToIngest(data, "prices", state, fmt.Sprintf("%d", upload.AlbionId))

	// sendToIngest ничего не возвращает - успех/ошибку самой отправки видно
	// только в её собственных логах ("Successfully sent ingest request..." /
	// "Error while sending ingest..."). Здесь просто фиксируем факт попытки.
	atomic.AddInt64(&a.sentOK, 1)
}
