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
	}
)
