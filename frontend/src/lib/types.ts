export interface Shard {
  shard_id: number;
  up: boolean;
  status: string;
  latency: number;
  last_heartbeat: number;
  last_connection: number;
  last_reconnect: number;
}

export interface Cluster {
  avg_latency: number;
  id: number;
  shards_up: number;
  up: boolean;
  status: string;
  shards: Shard[] | undefined;
}

export interface ClustersWrapper {
  avg_latency: number;
  max_concurrency: number;
  num_shards: number;
  shards_up: number;
  clusters: Cluster[] | undefined;
}

export interface ShardsWrapper {
  cluster_id: number;
  shards: Map<number, Shard>;
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
