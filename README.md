# kbot
DevOps application from scratch / DevOps застосунок з нуля
## Телеграм бот
Телеграм бот написано мовою Golang з метою ознайомлення з основними поняттями та функціями мови програмування.

Посилання на бота:  https://t.me/NickP_study_bot

### v1.0.6
Додано:

- автоматичний запуск Action Workflow при push до репозиторію у гілку main.

- Jenkins pipeline на мові groovy для білду артефакту за допомогою Jenkins

- виправлення та доопрацювання файлів Makefile та Dockerfile

### v1.0.5
Додано:

- створення Helm Chart для розгортання на Kubernetes кластер.

- автоматичний запуск Action Workflow при push до репозиторію у гілку develop.

[![Run Workflow](https://github.com/NickP007/kbot/actions/workflows/cicd-develop.yaml/badge.svg)](https://github.com/NickP007/kbot/actions/workflows/cicd-develop.yaml)


### v1.0.4
Додано команди:

/start - запуск бота та отримання первісної інструкції

/start hello, /hello або hello - привітання та вивід поточної версії бота

/help - допомога з використання бота

ping - Pong
