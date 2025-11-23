package spworlds

import "fmt"

func (c *Client) ClientCard() (*ClientCard, error) {
	var response ClientCard

	endpoint := "card"
	req, err := c.parseRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s request: %w", endpoint, err)
	}

	if err := c.doRequest(req, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", endpoint, err)
	}

	return &response, nil
}

func (c *Client) User(discordID string) (*User, error) {
	var response User

	endpoint := fmt.Sprintf("users/%s", discordID)
	req, err := c.parseRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s request: %w", endpoint, err)
	}

	if err := c.doRequest(req, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", endpoint, err)
	}

	return &response, nil
}

func (c *Client) UserCards(username string) (*[]Card, error) {
	var response []Card
	endpoint := fmt.Sprintf("accounts/%s/cards", username)
	req, err := c.parseRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s request: %w", endpoint, err)
	}

	if err := c.doRequest(req, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", endpoint, err)
	}

	return &response, nil
}

func (c *Client) Me() (*User, error) {
	var response User

	endpoint := "accounts/me"
	req, err := c.parseRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s request: %w", endpoint, err)
	}

	if err := c.doRequest(req, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", endpoint, err)
	}

	return &response, nil
}

type CreatePaymentOptions struct {
	// Товары транзакции.
	Items []Product `json:"items"`
	// URL страницы, на которую попадет пользователь после оплаты.
	RedirectURL string `json:"redirectUrl"`
	// URL, куда наш сервер направит запрос, чтобы оповестить ваш сервер об успешной оплате.
	WebhookUrl string `json:"webhookUrl"`
	// Cюда можно поместить любые полезные данных. Ограничение - 100 символов.
	// Data -> Payload
	Payload string `json:"data"`
}

func (c *Client) CreatePayment(opts CreatePaymentOptions) (*Payment, error) {
	var response Payment

	endpoint := "payments"
	req, err := c.parseRequest("POST", endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s request: %w", endpoint, err)
	}

	if err := c.doRequest(req, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", endpoint, err)
	}

	return &response, nil
}

type CreateTransactionOptions struct {
	// Номер карты получателя
	Receiver string `json:"receiver"`
	// Количество АРов для перевода
	Amount int `json:"amount"`
	// Комментарий для перевода
	Comment string `json:"comment"`
}

type CreateTransactionResponse struct {
	// Баланс карты после транзакции
	Balance int `json:"balance"`
}

func (c *Client) CreateTransaction(opts CreateTransactionOptions) (*CreateTransactionResponse, error) {
	var response CreateTransactionResponse

	endpoint := "transactions"
	req, err := c.parseRequest("POST", endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s request: %w", endpoint, err)
	}

	if err := c.doRequest(req, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", endpoint, err)
	}

	return &response, nil
}

type UpdateWebhookOptions struct {
	URL string `json:"url"`
}

type UpdateWebhookResponse struct {
	// Уникальный ID карты.
	ID string `json:"id"`
	// Обновленный webhook карты.
	Webhook string `json:"webhook"`
}

// На вебхук будут отправляться все новые транзакции связанные с картой.
// Данные будут отправлены через POST запрос.
func (c *Client) UpdateWebhook(opts UpdateWebhookOptions) (*UpdateWebhookResponse, error) {
	var response UpdateWebhookResponse

	endpoint := "card/webhook"
	req, err := c.parseRequest("POST", endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s request: %w", endpoint, err)
	}

	if err := c.doRequest(req, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", endpoint, err)
	}

	return &response, nil
}
