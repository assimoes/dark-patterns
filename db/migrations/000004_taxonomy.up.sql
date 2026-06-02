CREATE TABLE IF NOT EXISTS taxonomy_high_levels (
    id int generated always as identity primary key,
    name varchar(50) not null,
    description text not null,
    version int not null,
    active boolean not null default true,
    unique (name)
);

CREATE TABLE IF NOT EXISTS taxonomy_meso_levels (
    id int generated always as identity primary key,
    parent_id int not null references taxonomy_high_levels(id),
    code varchar(5) not null,
    name varchar(50) not null,
    description text not null,
    version int not null,
    active boolean not null default true,
    unique(code)
);