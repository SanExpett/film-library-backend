ALTER TABLE public."actor"
ADD COLUMN gender TEXT DEFAULT NULL CHECK (gender IN ('male', 'female', 'other'));