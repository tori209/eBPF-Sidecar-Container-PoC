CREATE TABLE "tasks" (
	"job_id" UUID NOT NULL,
	"task_id" UUID NOT NULL,
	"src_endpoint" VARCHAR NULL DEFAULT NULL,
	"src_bucket_name" VARCHAR NULL DEFAULT NULL,
	"src_object_name" VARCHAR NULL DEFAULT NULL,
	"dst_endpoint" VARCHAR NULL DEFAULT NULL,
	"dst_bucket_name" VARCHAR NULL DEFAULT NULL,
	"dst_object_name" VARCHAR NULL DEFAULT NULL,
	"range_begin" BIGINT NULL DEFAULT NULL,
	"range_end" BIGINT NULL DEFAULT NULL,
	"run_as_evil" BOOLEAN NULL DEFAULT NULL,
	PRIMARY KEY ("job_id", "task_id")
)
;
COMMENT ON COLUMN "tasks"."job_id" IS '';
COMMENT ON COLUMN "tasks"."task_id" IS '';
COMMENT ON COLUMN "tasks"."src_endpoint" IS '';
COMMENT ON COLUMN "tasks"."src_bucket_name" IS '';
COMMENT ON COLUMN "tasks"."src_object_name" IS '';
COMMENT ON COLUMN "tasks"."dst_endpoint" IS '';
COMMENT ON COLUMN "tasks"."dst_bucket_name" IS '';
COMMENT ON COLUMN "tasks"."dst_object_name" IS '';
COMMENT ON COLUMN "tasks"."range_begin" IS '';
COMMENT ON COLUMN "tasks"."range_end" IS '';
COMMENT ON COLUMN "tasks"."run_as_evil" IS '';
