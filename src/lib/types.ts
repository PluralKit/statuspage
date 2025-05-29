export interface Shard {
    shard_id: number;
    cluster_id: number;
    up: boolean;
    status: string;
    latency: number;
    last_heartbeat: Date;
    last_connection: Date;
}

export interface Cluster {
    cluster_id: number;
    avg_latency: number;
    up: boolean;
    status: string;
    shards: Shard[];
}

export interface IncidentUpdate {
    id: string;
    text: string;
    timestamp: Date;
}

export interface Incident {
    id: string;
    timestamp: Date;
    status: string;
    impact: string;
    updates: IncidentUpdate[];
    name: string;
    description: string;
    last_update: Date;
    resolution_timestamp: Date;
}

export interface Status {
    status: string;
    impact: string;
    active_incidents: string[];
    timestamp: Date;
}