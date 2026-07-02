# Пакет для взаимодействия с API серверов СП на Go

[Документация API](https://github.com/sp-worlds/api-docs)

## Быстрое начало

### Устрановка

```sh
go get github.com/xligenda/spworlds
```

### Использование

Для начала вам нужно указать ID и Token вашей карты, найти их можно [здесь](https://github.com/sp-worlds/api-docs/wiki/%D0%90%D1%83%D1%82%D0%B5%D0%BD%D1%82%D0%B8%D1%84%D0%B8%D0%BA%D0%B0%D1%86%D0%B8%D1%8F#%D0%BF%D0%BE%D0%BB%D1%83%D1%87%D0%B5%D0%BD%D0%B8%D0%B5-%D1%82%D0%BE%D0%BA%D0%B5%D0%BD%D0%B0-%D0%B8-id-%D0%BA%D0%B0%D1%80%D1%82%D1%8B)

```go
package main

import (
	"context"
	"fmt"

	"github.com/xligenda/spworlds"
)

func main() {
	api := spworlds.NewClient("card id", "card token", nil)

	resp, err := api.Me(context.Background())
	if err != nil || resp == nil {
		panic(err)
	}

	fmt.Printf("Никнейм владельца карточки: %s", resp.Username)
}
```

Перевод АРов на другую карту
```go
package main

import (
	"context"
	"fmt"

	"github.com/xligenda/spworlds"
)

func main() {
	api := spworlds.NewClient("card id", "card token", nil)

	// Перевод 10 АР на карту с номером OSTER, с комментарием "Подарок"
	resp, err := api.CreateTransaction(context.Background(),
		spworlds.
			NewCreateTransactionOptions("OSTER", 10).
			SetComment("Подарок"),
	)
	if err != nil || resp == nil {
		panic(err)
	}

	fmt.Printf("Оставшийся баланс карты: %d", resp.Balance)
}
```

Получение вебхука от SPWorlds
```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/xligenda/spworlds"
)

func main() {
	api := spworlds.NewClient("card id", "card token", nil)

	http.HandleFunc("/payment", func(w http.ResponseWriter, r *http.Request) {
		data, err := api.ParsePaymentDataValidated(r)
		if err != nil {
			log.Printf("Invalid payment webhook: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		fmt.Printf("Оплата от %s на сумму %d АР (payload: %s)\n", data.Payer, data.Amount, data.Payload)
		w.WriteHeader(http.StatusOK)
	})

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```