CREATE TABLE "logs" (
	"id" SERIAL NOT NULL,
	"job_id" UUID NULL DEFAULT NULL,
	"task_id" UUID NULL DEFAULT NULL,
	"timestamp" TIMESTAMPTZ NOT NULL,
	"src_ip" INET NULL DEFAULT NULL,
	"dst_ip" INET NULL DEFAULT NULL,
	"src_port" INTEGER NULL DEFAULT NULL,
	"dst_port" INTEGER NULL DEFAULT NULL,
	"packet_size" BIGINT NULL DEFAULT NULL,
	"protocol" VARCHAR NULL DEFAULT NULL,
	PRIMARY KEY ("id")
)
;
COMMENT ON COLUMN "logs"."id" IS '';
COMMENT ON COLUMN "logs"."job_id" IS '';
COMMENT ON COLUMN "logs"."task_id" IS '';
COMMENT ON COLUMN "logs"."timestamp" IS '';
COMMENT ON COLUMN "logs"."src_ip" IS '';
COMMENT ON COLUMN "logs"."dst_ip" IS '';
COMMENT ON COLUMN "logs"."src_port" IS '';
COMMENT ON COLUMN "logs"."dst_port" IS '';
COMMENT ON COLUMN "logs"."packet_size" IS '';
COMMENT ON COLUMN "logs"."protocol" IS '';
