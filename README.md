## 🚀 Технологический стек

* **Бэкенд:** Go (Golang)
* **База данных:** PostgreSQL
* **Брокер сообщений:** Apache Kafka
* **Контейнеризация:** Docker & Docker Compose
* **Фронтенд:** Чистый HTML / CSS / JavaScript
---
## ⚙️ Установка и запуск

Для запуска проекта на вашем компьютере должен быть установлен **Docker** и **Docker Compose**.

### Шаг 1: Клонирование репозитория

Склонируйте репозиторий на ваш локальный компьютер:
```bash
git clone [https://github.com/ВАШ_ЛОГИН/ВАШ_РЕПОЗИТОРИЙ.git]
cd ВАШ_РЕПОЗИТОРИЙ
```
### Шаг 2: Запуск сервисов
Выполните одну команду, чтобы собрать образ приложения и запустить все необходимые сервисы:

```Bash
docker-compose up --build
```
Дождитесь, пока все сервисы запустятся. Ваше приложение будет доступно по адресу http://localhost:8081.

## 🧪 Тестирование сервиса
Чтобы проверить работу системы, нужно сначала отправить данные в Kafka, а затем получить их через веб-интерфейс.

### Шаг 1: Отправка данных в Kafka
1. **Откройте новое окно терминала** (не закрывая то, в котором работает ```docker-compose```).

2. Запустите консольный продюсер Kafka, который подключится к нужному топику:

```Bash
docker-compose exec kafka kafka-console-producer.sh --bootstrap-server localhost:9092 --topic orders
```
3. Терминал перейдет в режим ожидания. Вставьте в него следующий JSON в одну строку и нажмите Enter:

```JSON

{"order_uid":"b563feb7b2b84b6test","track_number":"WBILMTESTTRACK","entry":"WBIL","delivery":{"name":"Test Testov","phone":"+9720000000","zip":"2639809","city":"Kiryat Mozkin","address":"Ploshad Mira 15","region":"Kraiot","email":"test@gmail.com"},"payment":{"transaction":"b563feb7b2b84b6test","request_id":"","currency":"USD","provider":"wbpay","amount":1817,"payment_dt":1637907727,"bank":"alpha","delivery_cost":1500,"goods_total":317,"custom_fee":0},"items":[{"chrt_id":9934930,"track_number":"WBILMTESTTRACK","price":453,"rid":"ab4219087a764ae0btest","name":"Mascaras","sale":30,"size":"0","total_price":317,"nm_id":2389212,"brand":"Vivienne Sabo","status":202}],"locale":"en","internal_signature":"","customer_id":"test","delivery_service":"meest","shardkey":"9","sm_id":99,"date_created":"2021-11-26T06:22:19Z","oof_shard":"1"}
```
4. В первом окне терминала (с логами docker-compose) вы должны увидеть сообщение от вашего приложения:

```Заказ b563feb7b2b84b6test успешно обработан и кэширован.```

### Шаг 2: Получение данных через веб-интерфейс
1. Откройте ваш браузер и перейдите по адресу: http://localhost:8081

2. В поле для ввода вставьте ```order_uid``` заказа: ```b563feb7b2b84b6test```

3. Нажмите кнопку "Найти".

На странице отобразятся отформатированные данные о вашем заказе.







