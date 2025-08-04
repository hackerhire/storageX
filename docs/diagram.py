from diagrams import Diagram, Cluster
from diagrams.onprem.client import Client
from diagrams.onprem.compute import Server
from diagrams.onprem.database import PostgreSQL
from diagrams.generic.database import SQL
from diagrams.generic.storage import Storage
from diagrams.onprem.aggregator import Fluentd

with Diagram("storageX System Architecture", show=False, direction="LR"):
    user = Client("User")
    cli = Server("CLI/Service")
    config = SQL("Config (JSON)")

    chunker = Server("Chunker")
    storage_svc = Server("StorageService (Orchestration)")
    manager = Server("StorageManager (Cloud Ops)")
    meta = PostgreSQL("MetadataService (SQLite)")

    with Cluster("Cloud Providers"):
        dropbox = Storage("Dropbox")
        gdrive = Storage("Google Drive")

    user >> cli
    cli >> storage_svc
    user >> config
    config >> storage_svc
    storage_svc >> chunker
    storage_svc >> manager
    storage_svc >> meta
    manager >> dropbox
    manager >> gdrive
    meta >> storage_svc
