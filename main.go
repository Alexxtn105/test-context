package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func main() {
	// Обычное использование контекста
	//res, err := testResponse()
	//if err != nil {
	//	log.Printf("Error: %v", err)
	//}

	//Дочерние контексты (по таймауту или дедлайну)
	//contextWithTimeout()  //контекст с тайм-аутом
	//contextWithDeadline() // тот же контекст с дедлайном

	// Контекст с отменой (каналы, WaitGroup)
	contextWithCancel()

}

// region Обычное использование контекста
func testResponse() (*http.Response, error) {
	// создаем контекст с таймаутом 15 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request with ctx: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform http request: %w", err)
	}

	return res, nil
}

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
		resultCh    = make(chan string)                        // Cюда будет записан результат
		ctx, cancel = context.WithCancel(context.Background()) // Создаем контекст с отменой
		services    = []string{"Uber", "Villagemobile", "Sett Taxi", "Yandex Go"}
		wg          sync.WaitGroup // Для упрощения работы с горутинами - фактически, это счетчик запущенных горутин.
		winner      string
	)

	// обязательно отменяем по окончании
	defer cancel()

	// запускаем несколько горутин
	for i := range services {
		svc := services[i]

		// Добавляем единичку в wg. Фактически это количество работающих горутин.
		// Увеличивать счетчик единицу нужно ДО запуска горутины.
		// Уменьшение счетчика должно производиться в исполняемой горутине с помощью метода wg.Done()
		// Когда счетчик wg станет равен 0, все заблокированные wg горутины продолжают выполнение (то есть, завершаются)
		// Если счетчик становится меньше 0 (такие случаи возможны), вызывается паника
		wg.Add(1)
		go func() {
			requestRide(ctx, svc, resultCh) // функция поиска такси
			wg.Done()                       // уменьшение счетчика горутин
		}()
	}

	go func() {
		winner = <-resultCh // Ожидаем, пока кто-нибудь предоставит нам машину
		cancel()            // Машина найдена, принудительно отменяем контекст. Все горутины с этим контекстом также завершат поиск
	}()
	// ожидаем, пока счетчик wg станет равен 0
	wg.Wait()
	log.Printf("found car in %q", winner)
}

func requestRide(ctx context.Context, serviceName string, resultCh chan string) {
	//time.Sleep(3 * time.Second)
	for {
		select {
		case <-ctx.Done(): // ожидаем из канала ответ Done()
			// контекст отменен
			log.Printf("stopped the search in %q (%v)", serviceName, ctx.Err())
			return
		default: // здесь с некоторой вероятностью пишем найденное имя сервиса (первое найденное такси)
			if rand.Float64() > 0.5 {
				// пишем в канал имя службы такси
				resultCh <- serviceName
				return
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
	}
}

//endregion
