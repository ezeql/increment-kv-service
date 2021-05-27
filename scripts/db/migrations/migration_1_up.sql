CREATE TABLE IF NOT EXISTS increments (
	"key" UUID NOT NULL,
	"value" BIGINT NOT NULL,
	CONSTRAINT increments_pk PRIMARY KEY ("key")
);