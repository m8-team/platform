COMPOSE=docker compose -f docker-compose.yml

.PHONY: up down restart logs shell ps clean

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

restart:
	$(COMPOSE) down
	$(COMPOSE) up -d

logs:
	$(COMPOSE) logs -f

shell:
	$(COMPOSE) exec app bash

ps:
	$(COMPOSE) ps

clean:
	$(COMPOSE) down -v