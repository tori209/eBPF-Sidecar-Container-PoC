CREATE TABLE "jobs" (
	"job_id" UUID NOT NULL,
	"src_endpoint" VARCHAR NULL DEFAULT NULL,
	"src_bucket_name" VARCHAR NULL DEFAULT NULL,
	"src_object_name" VARCHAR NULL DEFAULT NULL,
	"dst_endpoint" VARCHAR NULL DEFAULT NULL,
	"dst_bucket_name" VARCHAR NULL DEFAULT NULL,
	"dst_object_name" VARCHAR NULL DEFAULT NULL,
	"range_begin" BIGINT NULL DEFAULT NULL,
	"range_end" BIGINT NULL DEFAULT NULL,
	PRIMARY KEY ("job_id")
)
;
COMMENT ON COLUMN "jobs"."job_id" IS '';
COMMENT ON COLUMN "jobs"."src_endpoint" IS '';
COMMENT ON COLUMN "jobs"."src_bucket_name" IS '';
COMMENT ON COLUMN "jobs"."src_object_name" IS '';
COMMENT ON COLUMN "jobs"."dst_endpoint" IS '';
COMMENT ON COLUMN "jobs"."dst_bucket_name" IS '';
COMMENT ON COLUMN "jobs"."dst_object_name" IS '';
COMMENT ON COLUMN "jobs"."range_begin" IS '';
COMMENT ON COLUMN "jobs"."range_end" IS '';
