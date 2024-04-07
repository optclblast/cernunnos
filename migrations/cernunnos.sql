create type availability as enum (
	'available',
	'busy',
	'unavailable'
);

create table if not exists storages (
        id UUID primary key,
        name varchar(300),
        availability availability,
        created_at timestamp default current_timestamp,
        updated_at timestamp default current_timestamp
);

create table if not exists products (
        id UUID primary key,
        name varchar(300),
        size int default 0,
        created_at timestamp default current_timestamp,
        updated_at timestamp default current_timestamp
);

create table if not exists products_distribution (
        storage_id UUID references storages(id),
        product_id UUID references products(id),
        amount bigint default 0,
        reserved bigint default 0,
        available bigint default 0,
        created_at timestamp default current_timestamp,
        updated_at timestamp default current_timestamp,
        primary key (storage_id, product_id)
);

create table if not exists products_reservations (
        storage_id UUID references storages(id),
        product_id UUID references products(id),
        shipping_id UUID, 
        reserved bigint default 0,
        created_at timestamp default current_timestamp,
        updated_at timestamp default current_timestamp
);

create index if not exists index_products_reservations_storage_id_product_id_shipping_id
on products_reservations (
        product_id, shipping_id, storage_id
);

create index if not exists index_products_reservations_storage_id_product_id
on products_reservations (
        storage_id, product_id
);
	
create index if not exists index_products_reservations_storage_id_shipping_id
on products_reservations (
        storage_id, shipping_id
);

create index if not exists index_products_reservations_product_id_shipping_id
on products_reservations (
        product_id, shipping_id
);