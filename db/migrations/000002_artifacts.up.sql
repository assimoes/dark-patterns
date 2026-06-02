CREATE TABLE IF NOT EXISTS artifacts (
    id bigint generated always as identity primary key,
    modality varchar(20) not null check (modality in ('text', 'image', 'video')),
    source varchar(25) not null,
    source_id varchar(64),
    content_hash bytea not null,
    scraped_at timestamptz not null,
    created_at timestamptz not null default now(),
    unique (source, source_id),
    unique (modality, content_hash)
);

CREATE TABLE IF NOT EXISTS text_review_details (
    artifact_id bigint primary key references artifacts(id),
    body text not null,
    voted_up boolean not null,
    hours_played int not null,
    lang varchar(25) not null
);

CREATE TABLE IF NOT EXISTS image_details (
    artifact_id bigint primary key references artifacts(id),
    image_uri text not null,
    width int,
    height int,
    mime_type varchar(30),
    ocr_text text
);
