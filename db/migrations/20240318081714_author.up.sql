ALTER TABLE public."actor"
    ADD COLUMN author_id  BIGINT NOT NULL REFERENCES public."user" (id);
