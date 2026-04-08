Для вызова API сервисов платформы используйте access token, выданный вашим
провайдером идентификации.

Обычно access token получают одним из следующих способов:

- для пользовательского аккаунта через интерактивный вход;
- для сервисного аккаунта через machine-to-machine сценарий, например
  `client_credentials`;
- для федеративного аккаунта через внешний identity provider и последующий
  обмен на access token.

Передавайте полученный токен в заголовке `Authorization` при обращении к API:

```http
Authorization: Bearer <access_token>
```

Если токен сохранен в переменной окружения, используйте ее:

```http
Authorization: Bearer ${ACCESS_TOKEN}
```

Для gRPC-вызовов передавайте то же значение в metadata `authorization`.
