package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

func main() {

	//region //Обычное использование контекста
	//res, err := testResponse()
	//if err != nil {
	//	log.Printf("Error: %v", err)
	//}
	//endregion

	//region Дочерние контексты (по таймауту или дедлайну)

	fmt.Println("-----------------Дочерние контексты (по таймауту или дедлайну)-------------------")
	contextWithTimeout()  //контекст с тайм-аутом
	contextWithDeadline() // тот же контекст с дедлайном

	//endregion

	//region Контекст с отменой (каналы, WaitGroup)

	fmt.Println("-----------------Контекст с отменой (каналы, WaitGroup)-------------------")
	contextWithCancel()

	//endregion

	//region Контекст с параметрами

	fmt.Println("-----------------Контекст с параметрами-------------------")
	contextWithValue()

	//endregion

	//region WaitGroup

	fmt.Println("-----------------WaitGroup-------------------")
	wg := &sync.WaitGroup{}
	cs := make(chan string)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go worker(wg, cs, i)
	}

	go monitorWorker(wg, cs)

	done := make(chan bool, 1)
	go printWorker(cs, done)
	fmt.Println(<-done)

	//endregion
}

// region //Обычное использование контекста

//func testResponse() (*http.Response, error) {
//	// создаем контекст с таймаутом 15 секунд
//	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
//	defer cancel()
//
//	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com", nil)
//	if err != nil {
//		return nil, fmt.Errorf("failed to create request with ctx: %w", err)
//	}
//
//	res, err := http.DefaultClient.Do(req)
//	if err != nil {
//		return nil, fmt.Errorf("failed to perform http request: %w", err)
//	}
//
//	return res, nil
//}

//endregion

// region Дочерние контексты (по таймауту или дедлайну)
func contextWithTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	doWork(ctx)
}

func contextWithDeadline() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Second))
	defer cancel()
	doWork(ctx)
}

func doWork(ctx context.Context) {
	// создаем дочерний контекст
	newCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log.Println("Start working...")

	for {
		select {
		case <-newCtx.Done():
			log.Printf("ctx done: %v", ctx.Err())
			return
		default:
			log.Printf("working...")
			time.Sleep(1 * time.Second)
		}
	}
}

//endregion

// region Контекст с отменой (каналы, WaitGroup)
func contextWithCancel() {

	// Задача: найти такси как можно быстрее.
	// Осуществляем поиск в нескольких сервисах одновременно
	// Если такси найдено - отменить поиск в оставшихся сервисах
	var (
		services = []string{"Uber", "Пчёлка", "Maxim", "Yandex Go"}

		// Создаем буферизированный канал для результата размером по количеству служб
		// Если использовать небуферизированный канал вида resultCh = make(chan string),
		// то будут возникать deadlocks из-за того, что горутины будут пытаться записать данные в канал, данные из которого еще не прочитаны
		resultCh    = make(chan string, len(services))         // Здесь будет записан результат
		ctx, cancel = context.WithCancel(context.Background()) // Создаем ДОЧЕРНИЙ контекст с отменой
		wg          sync.WaitGroup                             // Для упрощения работы с горутинами - фактически, это счетчик запущенных горутин.
		winner      string                                     // Имя сервиса-победителя
	)

	// обязательно отменяем по окончании
	defer cancel()

	// Запускаем несколько горутин.
	// Перед запуском добавляем 1 в wg.Add(1) - фактически это количество работающих горутин.
	// Увеличивать счетчик единицу нужно ДО ЗАПУСКА горутины.
	// Уменьшение счетчика должно производиться В ИСПОЛНЯЕМОЙ ГОРУТИНЕ с помощью метода wg.Done()
	// Когда счетчик wg станет равен 0, все заблокированные wg горутины продолжают выполнение (то есть, завершаются)
	// Если счетчик становится меньше 0 (такие случаи возможны), вызывается паника
	for i := range services {
		svc := services[i]
		// инкрементим wg на 1 ДО запуска горутины
		wg.Add(1)
		go func() {
			defer wg.Done()                 // уменьшение счетчика горутин
			requestRide(ctx, svc, resultCh) // функция поиска такси
		}()
	}
	// в отдельной горутине мониторим канал получения результата
	go func() {
		winner = <-resultCh // Ожидаем, пока кто-нибудь предоставит нам машину
		cancel()            // Машина найдена, принудительно отменяем контекст. Все горутины с этим контекстом также завершат поиск
	}()

	// ожидаем, пока счетчик wg станет равен 0
	wg.Wait()
	// на всякий случай закрываем канал
	close(resultCh)
	log.Printf("found car in %q", winner)
}

func requestRide(ctx context.Context, serviceName string, resultCh chan string) {
	log.Printf("searching in %q...", serviceName)
	time.Sleep(1 * time.Second)
	for {
		select {
		case <-ctx.Done(): // ожидаем из канала ответ Done()
			// контекст отменен
			log.Printf("stopped the search in %q (%v)", serviceName, ctx.Err())
			return
		default: // здесь с некоторой вероятностью пишем найденное имя сервиса (первое найденное такси)
			if rand.Float64() > 0.95 {
				// пишем в канал имя службы такси
				resultCh <- serviceName
				fmt.Println("MATCHED:", serviceName)
				return
			}
			continue
		}
	}
}

//endregion

// region Контекст с параметрами
func contextWithValue() {
	// Данные через контекст лучше не передавать - это АНТИПАТТЕРН!!!
	// Потому что это порождает неявный контракт между компонентами приложения, он ненадежен!!!
	// за исключением предоставления внешней библиотеке нашей реализации интерфейса, который мы не можем менять
	// например, middleware в http-функциях
	// во всех остальных случаях лучше применять аргументы функции

	ctx := context.WithValue(context.Background(), "name", "Joe")

	log.Printf("name = %v", ctx.Value("name"))
	log.Printf("age = %v", ctx.Value("age")) // вывод nil (ключа не существует)

}

//endregion

// region WaitGroup
func worker(wg *sync.WaitGroup, cs chan string, i int) {
	defer wg.Done()
	cs <- "worker" + strconv.Itoa(i)
}

func monitorWorker(wg *sync.WaitGroup, cs chan string) {
	wg.Wait()
	close(cs)
}

func printWorker(cs <-chan string, done chan<- bool) {
	for i := range cs {
		fmt.Println(i)
	}

	done <- true
}

//endregion
