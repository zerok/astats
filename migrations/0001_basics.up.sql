create table if not exists urls (
    id integer primary key autoincrement,
    url text not null unique
);

create table if not exists requests_per_day (
    url_id integer references urls(id) on delete cascade,
    date text not null,
    count int not null
);
