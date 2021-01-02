package sqlite

var (
	migrations = inlineMigration{
		`
-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied

create table if not exists items
(
    id          integer 
        constraint items_pk 
            primary key autoincrement,
    user_id     integer,
    name        varchar [80],
    description varchar [250] default '',
    url         varchar [200] default '',
    image       varchar [200] default '',
    rank        integer       default 0,
    compared    integer       default 0
);

create unique index if not exists items_user_id_name_uindex on items (user_id, name);

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back

drop index items_user_id_name_uindex;

drop table items;
`,
		`
-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied

create table categories (
	id          integer 
        constraint items_pk 
            primary key autoincrement,
    user_id     integer,
    name varchar [80],
    description varchar [250] default ''
);

create unique index if not exists categories_user_id_name_uindex on categories (user_id, name);

alter table items add column category integer default 0;

drop index items_user_id_name_uindex;

create unique index if not exists items_user_id_name_cat_uindex on items (user_id, name, category);
create index items_user_id_category_index
	on items (user_id, category);

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back

create unique index if not exists items_user_id_name_uindex on items (user_id, name);

drop index items_user_id_name_cat_uindex;

drop index items_user_id_category_index
drop index categories_user_id_name_uindex;

drop table categories;
`,
		`
-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied

create table users (
	id          integer 
        constraint items_pk 
            primary key,
    config TEXT default ''
);

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back

drop table users;

`,
	}
)
