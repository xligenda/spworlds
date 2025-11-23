package spworlds

type ClientCard struct {
	Balance int     `json:"balance"`
	Webhook *string `json:"webhook"`
}

type User struct {
	// Уникальный ID аккаунта
	ID *string `json:"id"`
	// Ник пользователя или nill, если у пользователя нет входа на сервер
	Username *string `json:"username"`
	// Minecraft UUID пользователя или nill, если у пользователя нет входа на сервер
	// Если это Bedrock сервер, то данного поля не будет в запросе
	UUID *string `json:"uuid"`
	// Статус игрока
	Status *string `json:"status"`
	// Массив содержащий роли аккаунта
	Roles *[]string `json:"roles"`
	// Информация о городе, в котором состоит игрок Если игрок не состоит в городе, вернется nill
	Cities []*CityMember `json:"cities"`
	// Массив содержащий карты игрока
	Cards *[]Card `json:"cards"`
	// Дата создания аккаунта
	CreatedAt *Timestamp `json:"createdAt"`
}

type CityMember struct {
	// Должность в городе
	Role CityMemberRole `json:"role"`
	// Дата создания города
	CreatedAt Timestamp `json:"createdAt"`
	City      City      `json:"city"`
}

func (m CityMember) IsMayor() bool {
	return m.Role == Mayor
}

type City struct {
	// Уникальный ID города
	ID string `json:"id"`
	// Название города
	Name string `json:"name"`
	// X координата города в верхнем мире
	OverX int `json:"x"`
	// Z координата города в верхнем мире
	OverZ int `json:"z"`
	// X координата города в нижнем мире
	NetherX int `json:"netherX"`
	// Z координата города в нижнем мире
	NetherZ int `json:"netherZ"`
	// Цвет ветки, на которой зарегистрирован город
	Lane LaneColor `json:"lane"`
}

type Card struct {
	// Уникальный ID карты
	ID *string `json:"id"`
	// Название карты
	Name string `json:"name"`
	// Номер карты
	Number string `json:"number"`
	// Цвет карты
	Color *CardColor `json:"color"`
}

type Product struct {
	// Название товара
	// Ограничение: 3-32 символа
	Name string `json:"name"`
	// Кол-во данного товара
	// Ограничение: 1-9999
	Count int `json:"count"`
	// Цена за 1 ед. товара
	// Ограничение: 1-1728
	Price int `json:"price"`
	// Не обязательный параметр
	// Комментарий к товару
	// Ограничение: 3-64 символа
	Comment string `json:"comment"`
}

type Payment struct {
	// Название карты получателя
	Card string `json:"card"`
	// Код для оплаты транзакции
	Code string `json:"code"`
	// Ссылка на страницу оплаты, на которую стоит перенаправить пользователя.
	URL string `json:"url"`
}
