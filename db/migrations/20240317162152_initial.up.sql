CREATE SEQUENCE IF NOT EXISTS user_id_seq;
CREATE SEQUENCE IF NOT EXISTS film_id_seq;
CREATE SEQUENCE IF NOT EXISTS actor_id_seq;
CREATE SEQUENCE IF NOT EXISTS film_actor_id_seq;


CREATE TABLE IF NOT EXISTS public."user"
(
    id         BIGINT                   DEFAULT NEXTVAL('user_id_seq'::regclass) NOT NULL PRIMARY KEY,
    email      TEXT UNIQUE                                                       NOT NULL CHECK (email <> '')
    CONSTRAINT max_len_email CHECK (LENGTH(email) <= 256),
    password   TEXT                                                              NOT NULL CHECK (password <> '')
    CONSTRAINT max_len_password CHECK (LENGTH(password) <= 256),
    is_admin   BOOL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()                            NOT NULL
);

CREATE TABLE IF NOT EXISTS public."film"
(
    id           BIGINT                   DEFAULT NEXTVAL('film_id_seq'::regclass)    NOT NULL PRIMARY KEY,
    author_id    BIGINT                                                               NOT NULL REFERENCES public."user" (id),
    title        TEXT                                                                 NOT NULL CHECK (title <> '')
    CONSTRAINT   max_len_title CHECK (LENGTH(title) <= 150),
    description  TEXT                                                                 NOT NULL CHECK (description <> '')
    CONSTRAINT   max_len_description CHECK (LENGTH(description) <= 1000),
    release_date TIMESTAMP WITH TIME ZONE,
    rating       INTEGER NOT NULL
    CONSTRAINT   from_zero_to_ten_rating CHECK (rating >= 0 and rating <= 10),
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW()                               NOT NULL
);

CREATE TABLE IF NOT EXISTS public."actor"
(
    id         BIGINT                   DEFAULT NEXTVAL('actor_id_seq'::regclass) NOT NULL PRIMARY KEY,
    name       TEXT UNIQUE DEFAULT NULL
    CONSTRAINT max_len_name CHECK (LENGTH(name) <= 256),
    birthday   TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()                            NOT NULL
);

CREATE TABLE IF NOT EXISTS public."film_actor"
(
    id        BIGINT                   DEFAULT NEXTVAL('film_actor_id_seq'::regclass) NOT NULL PRIMARY KEY,
    film_id   BIGINT NOT NULL REFERENCES public."film" (id),
    actor_id  BIGINT NOT NULL REFERENCES public."actor" (id)
);
