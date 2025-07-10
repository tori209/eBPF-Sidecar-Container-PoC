import pandas as pd
import uuid

case_name = "anomaly_2"
uuid_cols_name = ["job_id", "task_id"]

def conv_uuid(target):
    global uuid_cols_name
    df = pd.read_parquet(f'{target}.parquet', engine='pyarrow')
    for col in df.columns:
        if col in uuid_cols_name:
            df[col] = df[col].apply(lambda x: uuid.UUID(bytes=x))
    return df

def unlimit_print():
    pd.set_option('display.max_rows', None)
    pd.set_option('display.max_columns', None)
    pd.set_option('display.width', None)
    pd.set_option('display.max_colwidth', None)


def limit_print():
    pd.reset_option('display.max_rows')
    pd.reset_option('display.max_columns')
    pd.reset_option('display.width')
    pd.reset_option('display.max_colwidth')

unlimit_print()

log_df = conv_uuid(f"./{case_name}/logs")
task_df = conv_uuid(f"./{case_name}/tasks")
tl_df = pd.merge(log_df, task_df, on=['job_id','task_id'], how='inner')[['job_id','task_id','src_endpoint','dst_endpoint','range_begin','range_end','src_ip','src_port','dst_ip','dst_port','packet_size','timestamp','run_as_evil']]
#display(log_df.head(1))
print(f"total log count: {len(log_df)}")
tl_df = tl_df[
    (tl_df["job_id"] != uuid.UUID("00000000-0000-0000-0000-000000000000")) & 
    (tl_df["dst_port"] != 53) &
    (tl_df["src_port"] != 8080) # RPC Response
]
print(tl_df.groupby(["job_id", "task_id", "run_as_evil"]).agg({
    "timestamp": (lambda x : max(x) - min(x)),
    "dst_ip": pd.Series.nunique,
}))
#print("=" * 50)

#log_df = log_df[
#    (log_df["job_id"] != uuid.UUID("00000000-0000-0000-0000-000000000000")) & 
#    (log_df["dst_port"] != 53)]
#print(log_df[["dst_ip", "dst_port"]].drop_duplicates(["dst_ip", "dst_port"]))
#print(log_df.groupby(["dst_ip"])["dst_ip"].agg("count"))

limit_print()