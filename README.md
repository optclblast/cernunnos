# Cernunnos. Сервис управления складом
## Сбока проекта и запуск
В корне проекта выполните команду  
``` bash
make up
```
Эта команда соберет docker контейнер и поднимет compose проект. А так же наполнит бд случайными данными для тестов.    
Если нужно остановить compose проект, выполните команду
``` bash
make down
```
Для заполнения БД данными для тестов в корне выполните команду   
``` bash
make filldb
```

## Тестирование и линтер
Для запуска линтера в корне выполните команду    
``` bash
make tools # Установит линтер в папку bin/tools   
make lint # Запустит линтер   
```

Для запуска тестов разверните проект по инструкции выше. Затем исполните команду 
``` bash
make test
```

## API

Формат дат в ответе - unix milli.   

### Получение списка продуктов    
Эндпоинт **\[GET\] /products**    
Пример запроса:    
```bash
curl --location --request GET 'http://localhost:8080/products' \
--header 'Content-Type: application/json' \
--data '{
    "ids":[
        "7baeb7d9-32d5-42ac-b1ec-86d134d73e93",
        "8816f991-a041-480f-9286-8c9e737e1539",
        "bad6c6c4-8f62-4b8b-b4dc-5424fc95c4dc"
    ],
    "storage_id":"34152f06-bb83-4566-9bb8-68abf3dd4560",
    "limit":25,
    "offset":25
}'
```
Параметры:   
1. ids | type:strings-array \[optional\]   
Если передан, то выборка будет производится только по перечисленным товарам.   
2. storage_id | type:string \[optional\]    
Если передан, выборка будет производится только в определенном складе   
3. limit | type:int \[optional\]    
Параметр позволяет ограничить размер коллекции в ответе. Максимум элементов - 500.   
3. offset | type:int \[optional\]    
Параметр, предназначенный для пагинации. Поскольку размер выборки ограничен 500 элементами, мы можем отправить несколько запросов (если нужно), передав в кажлм последующем offset из ответа   
4. with_unavailable | type:bool  \[optional\]   
Если **НЕ** передан и передан id склада, в ответ придут только те продукты, которые не зарезервированы полностью (available > 0)
   
Пример ответа:   
```json
{
    "products": [
        {
            "id": "bad6c6c4-8f62-4b8b-b4dc-5424fc95c4dc",
            "name": "Gray Lightbulb Elite",
            "size": 232,
            "created_at": 1712604303630,
            "updated_at": 1712604303630
        },
        {
            "id": "7baeb7d9-32d5-42ac-b1ec-86d134d73e93",
            "name": "Maroon Leather Drone",
            "size": 46,
            "created_at": 1712604303632,
            "updated_at": 1712604303632
        },
        {
            "id": "8816f991-a041-480f-9286-8c9e737e1539",
            "name": "Navy Granite Scale",
            "size": 37,
            "created_at": 1712604303866,
            "updated_at": 1712604303866
        }
    ],
    "offset": 3
}
```

### Получение списка складов     
Эндпоинт **\[GET\] /storages**     
Пример запроса:    
``` bash
curl --location --request GET 'localhost:8081/storages' \
--header 'Content-Type: application/json' \
--data '{
    "ids":[
        "db434e41-b1cc-4f88-b804-83a66e024db2"
    ],
    "limit":25,
    "offset":0
}'
```
Параметры:    
1. ids | type:strings-array \[optional\]   
Если передан, то выборка будет производится только по перечисленным складам.   
2. limit | type:int \[optional\]    
Параметр позволяет ограничить размер коллекции в ответе. Максимум элементов - 500.   
3. offset | type:int \[optional\]    
Параметр, предназначенный для пагинации. Поскольку размер выборки ограничен 500 элементами, мы можем отправить несколько запросов (если нужно), передав в каждом последующем offset из ответа   

Пример ответа:   
```json 
{
    "storages": [
        {
            "id": "db434e41-b1cc-4f88-b804-83a66e024db2",
            "name": "Maudite",
            "reserved": 28117,
            "available": 87776,
            "created_at": 1712690332997,
            "updated_at": 1712690332997
        }
    ],
    "offset": 1
}
```

### Получение списка товаров на конкретном складе      
Эндпоинт **\[GET\] /storages/{storage_id}/products**    
Пример запроса:    
``` bash
curl --location --request GET 'http://localhost:8080/storages/db434e41-b1cc-4f88-b804-83a66e024db2/products' \
--header 'Content-Type: application/json' \
--data '{
    "ids": [
        "25937bb3-d77f-45f9-ab92-c955dbe71c78",
        "a3d0292e-d0be-4292-857f-4e9b7cd825c4"
    ],
    "with_unavailable": true,
    "limit": 25,
    "offset": 0
}'
```
Параметры:   
1. ids | type:strings-array \[optional\]   
Если передан, то выборка будет производится только по перечисленным товарам.     
2. limit | type:int \[optional\]    
Параметр позволяет ограничить размер коллекции в ответе. Максимум элементов - 500.   
3. offset | type:int \[optional\]    
Параметр, предназначенный для пагинации. Поскольку размер выборки ограничен 500 элементами, мы можем отправить несколько запросов (если нужно), передав в каждом последующем offset из ответа   
4. with_unavailable | type:bool  \[optional\]   
Если **НЕ** передан, в ответ придут только те продукты, которые не зарезервированы полностью (available > 0)

Пример ответа:
```json
{
    "products": [
        {
            "id": "25937bb3-d77f-45f9-ab92-c955dbe71c78",
            "name": "Microwave Bright Advanced",
            "size": 173,
            "created_at": 1712690332988,
            "updated_at": 1712690332988,
            "storage_id": "db434e41-b1cc-4f88-b804-83a66e024db2",
            "amount": 8728,
            "reserved": 3843,
            "available": 4885
        },
        {
            "id": "a3d0292e-d0be-4292-857f-4e9b7cd825c4",
            "name": "Eco-Friendly Gadget Innovative",
            "size": 247,
            "created_at": 1712690333004,
            "updated_at": 1712690333004,
            "storage_id": "db434e41-b1cc-4f88-b804-83a66e024db2",
            "amount": 3111,
            "reserved": 288,
            "available": 2823
        }
    ],
    "offset": 2
}
```

### Получение списка резервов продуктов для доставки   
Эндпоинт **\[GET\] /reservations**  
Пример запроса:   
```bash
curl --location --request GET 'http://localhost:8080/reservations' \
--header 'Content-Type: application/json' \
--data '{
    "storage_id":"d910311b-b77c-48a2-be38-8e4b301e9de2",
    "product_id":"d6dc4546-7663-4d1d-ba28-dddb04b49053",
    "shipping_id":"c2ecb8dc-32b7-4cd4-b653-de8d87e6423f",
    "limit":25,
    "offset":0
}'
```
Параметры:   
1. storage_id | type:string \[required\]   
На каком складе зарезервирован товар.
2. product_id | type:string \[optional\]    
Если передан, выборка будет производится только по конектному товару.
3. shipping_id | type:string \[required\]   
На какую доставку зарезервирован товар.
4. limit | type:int \[optional\]    
Параметр позволяет ограничить размер коллекции в ответе. Максимум элекментов - 500.   
5. offset | type:int \[optional\]    
Параметр, предназначенный для пагинации. Поскольку размер выборки ограничен 500 элементами, мы можем отправить несколько запросов (если нужно), передав в кажлм последующем offset из ответа   
   
Пример ответа:   
```json
{
    "reservations": [
        {
            "storage_id": "d910311b-b77c-48a2-be38-8e4b301e9de2",
            "product_id": "d6dc4546-7663-4d1d-ba28-dddb04b49053",
            "shipping_id": "c2ecb8dc-32b7-4cd4-b653-de8d87e6423f",
            "reserved": 2846,
            "created_at": 1712604303979,
            "updated_at": 1712604303979
        }
    ],
    "offset": 1
}
```

### Создание резерва продуктов для доставки на складе    
Эндпоинт **\[POST\] /reservations/new**   
Пример запроса:   
```bash
curl --location 'http://localhost:8080/reservations/new' \
--header 'Content-Type: application/json' \
--data '{
    "products": [
        "d6dc4546-7663-4d1d-ba28-dddb04b49053"
    ],
    "storage_id":"d910311b-b77c-48a2-be38-8e4b301e9de2",
    "shipping_id": "c2ecb8dc-32b7-4cd4-b653-de8d87e6423f",
    "amount": 10
}'
```
Параметры:   
1. storage_id | type:string \[optional\]   
На каком складе зарезервирован товар. Если не передан, резервы автоматически распределятся по складам, если на одном не будет хватать места
2. products | type:strings-array \[required\]    
Какие товары необходимо зарезервировать
3. shipping_id | type:string \[required\]   
На какую доставку зарезервирован товар.
4. amount | type:int \[required\]    
Кол-во товаров для резервирования


Пример ответа:   
```json
{
    "ok": true
}
```

Пример ошибки:
```json
{
    "code": 400,
    "details": "Not All Required Fields Provided! See API Documentation for more info"
}
```

```json
{
    "code": 507,
    "details": "Not Enough Space In Storage(s)!"
}
```


### Отмена резерва продуктов для доставки на складе    
Эндпоинт **\[DELETE\] /reservations/cancel**    
Пример запроса:    
```bash
curl --location --request DELETE 'http://localhost:8080/reservations/cancel' \
--header 'Content-Type: application/json' \
--data '{
    "products": [
        "d6dc4546-7663-4d1d-ba28-dddb04b49053"
    ],
    "storage_id":"d910311b-b77c-48a2-be38-8e4b301e9de2",
    "shipping_id": "c2ecb8dc-32b7-4cd4-b653-de8d87e6423f"
}'
```
Параметры:   
1. storage_id | type:string \[optional\]   
На каком складе зарезервирован товар. Если не передан, будут отменены резервы на всех складах для этих товаров для этой доставки
2. products | type:strings-array \[required\]    
Резерв каких товаров нужно отменить
3. shipping_id | type:string \[required\]   
На какую доставку зарезервирован товар.
   
Резерв товара будет отменен и товары будут снова доступны на складе

Пример ответа:   
```json
{
    "ok": true
}
```

Пример ошибки:
```json
{
    "code": 400,
    "details": "Not All Required Fields Provided! See API Documentation for more info"
}
```

### Списание зарезервированных товаров со склада    
Эндпоинт **\[DELETE\] /reservations/release**    
Пример запроса:    
```bash
curl --location --request DELETE 'http://localhost:8080/reservations/release' \
--header 'Content-Type: application/json' \
--data '{
    "products": [
        "d6dc4546-7663-4d1d-ba28-dddb04b49053"
    ],
    "storage_id":"d910311b-b77c-48a2-be38-8e4b301e9de2",
    "shipping_id": "c2ecb8dc-32b7-4cd4-b653-de8d87e6423f"
}'
```
Параметры:   
1. storage_id | type:string \[optional\]   
На каком складе зарезервирован товар. Если не передан, будут списаны резервы на всех складах для этих товаров для этой доставки
2. products | type:strings-array \[required\]    
Какие товары нужно списать
3. shipping_id | type:string \[required\]   
На какую доставку зарезервирован товар.
   
Резерв товара будет отменен и товары будут снова доступны на складе

Пример ответа:   
```json
{
    "ok": true
}
```

Пример ошибки:
```json
{
    "code": 400,
    "details": "Not All Required Fields Provided! See API Documentation for more info"
}
```
