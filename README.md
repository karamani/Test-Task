# Test-Task
Выполненное тестовое задание для NodaSoft.
createtable.sql — файл с запросом на создание БД и таблицы
pricelist.csv — тестовый csv файл прайс-листа с данными
NSTest/main.go — исходный код веб-службы на golang
index.html — страница для отправки запросов

Веб-служба запускается с параметром -conn string, где string - строка, содержащая параметры соединения с БД в формате:
  username:password@protocol(host:port)/database
  
По умолчанию используется бесплатный сервер mySQL.
