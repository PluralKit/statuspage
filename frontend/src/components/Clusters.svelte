<script lang="ts">
    import { onMount } from 'svelte';
    import { slide } from 'svelte/transition';
    
    import { dateAgo, getShardID } from '$lib/util';
    import { PUBLIC_SHARD_URL } from '$env/static/public';
    import { type Cluster, type Shard } from '$lib/types';

    let clusters: Cluster[] = [];
    let shards_up = 0;
    let shards_total = 0;
    let avg_latency = 0;
    let max_concurrency = 16;
    let error: any = undefined;

    //kinda a janky fix for closing, but whatevs it'll work for now
    function clickHandler(event: MouseEvent){
        const clicked = event.target as HTMLElement;
        if(clicked === null || ! clicked.tagName) return;
        if(clicked.tagName !== 'BUTTON' && clicked.tagName !== 'INPUT' && clicked.tagName !== 'A' 
            && !clicked.classList.contains("cluster") && !clicked.classList.contains("shard") && !clicked.classList.contains("btn")){
            if(!findCluster) showCluster = false;
        }
    }

    onMount(async () => {
        try {
            const response = await fetch(PUBLIC_SHARD_URL);
            const data = await response.json();

            let shards: Shard[] = data.shards.map((item: any) => {
                return {
                    shard_id: item.shard_id,
                    cluster_id: item.cluster_id,
                    up: item.up,
                    latency: item.latency,
                    last_heartbeat: new Date(item.last_heartbeat * 1000),
                    last_connection: new Date(item.last_connection * 1000),
                } as Shard;
            });

            shards.sort((a, b) => a.shard_id - b.shard_id);

            shards.forEach((s) => {
                if(!clusters[s.cluster_id]) clusters[s.cluster_id] = <Cluster>{cluster_id: s.cluster_id, avg_latency: 0, up: true, status: "healthy", shards: []};
                clusters[s.cluster_id].shards.push(s);
                if(s.up){
                    shards_up++;
                    avg_latency += s.latency;
                }
            });
            shards_total = shards.length;
            avg_latency = Math.floor(avg_latency/shards_up);
            
            clusters.forEach((c)=>{
                let l = 0;
                c.shards.forEach((s)=>{
                    l+=s.latency;
                    
                    if(!s.up) s.status = "down";
                    else if (s.latency < 300) s.status = "healthy";
                    else if (s.latency < 600) s.status = "degraded";
                    else s.status = "severe";
                })
                c.avg_latency = Math.floor(l/c.shards.length);
                
                if(!c.up) c.status = "down";
                else if (c.avg_latency < 300) c.status = "healthy";
                else if (c.avg_latency < 600) c.status = "degraded";
                else c.status = "severe";
            })

            max_concurrency = clusters[0].shards.length;
        } catch (e) {
            error = e;
            console.error(e);
        }
    });

    let findClusterInput = "";
    let findClusterErr = "";
    let shownCluster: Cluster;
    let shownShardID: number;
    let showCluster = false;
    let findCluster = false;

    //TODO: clean this up hehe, better validation?
    function clusterInfoHandler() {
        if(findClusterInput == "") {
            showCluster = false;
            findCluster = false;
            findClusterErr = "";
            shownCluster = null as any;
            shownShardID = -1;
            return;
        }

        var match = findClusterInput.match(/https:\/\/(?:[\w]*\.)?discord(?:app)?\.com\/channels\/(\d+)\/\d+\/\d+/);
        if(match != null) {
            let shardID = getShardID(match[1], shards_total);
            if (shardID != -1){
                shownShardID = Number(shardID);
                shownCluster = clusters[Math.floor(shownShardID / max_concurrency)];
                showCluster = true;
                findCluster = true;
                findClusterErr = "";
                return;
            }
        }

        try {
            var shardID = getShardID(findClusterInput, shards_total);
            if(shardID == -1 || !shardID) throw new Error();
            shownShardID = Number(shardID);
            shownCluster = clusters[Math.floor(shownShardID / max_concurrency)];
            showCluster = true;
            findCluster = true;
            findClusterErr = "";
            return;
        } catch(e) {
            showCluster = false;
            findCluster = false;
            findClusterErr = "Invalid server ID";
        }
    }
    function showClusterHandler(id: number) {
        if(showCluster && id === shownCluster.cluster_id) {
            showCluster = false;
        } else {
            findClusterInput = "";
            findCluster = false;
            shownCluster = clusters[id];
            showCluster = true;
        }
    }
</script>

<svelte:body onclick={clickHandler} />

<div class="card bg-base-200 shadow-sm">
    <div class="p-8 flex flex-col gap-4 w-full">
        {#if error}
        <div role="alert" class="alert alert-error">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 shrink-0 stroke-current" fill="none" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span>An error occured while fetching status data: {error.message}</span>
          </div>
        {/if}
        <h2 class="text-lg">Cluster Status:</h2>
        <div class="stats bg-base-100 shadow stats-vertical sm:stats-horizontal" role="region" aria-label="Overall statistics">
            <div class="stat">
              <div class="stat-title">Shards Up</div>
              <div class="stat-value">{shards_up} / {shards_total}</div>
            </div>
            <div class="stat">
                <div class="stat-title"> Average Latency</div>
                <div class="stat-value">{avg_latency} ms</div>
            </div>
        </div>
        <div class="flex flex-col items-center w-full" role="region" aria-label="Cluster status">
            <div class="cluster-ctr flex flex-wrap flex-row py-6 justify-start">
                {#each clusters as cluster}
                <button class="cluster aspect-square tooltip indicator {cluster.status}" on:click={()=>{showClusterHandler(cluster.cluster_id)}}>
                    {cluster.cluster_id}
                    <div class="tooltip-content">
                        avg latency: {cluster.avg_latency}
                    </div>
                </button>
                {/each}
            </div>
        </div>

        {#if showCluster}
        <div class="card bg-base-100 py-8 px-2" transition:slide="{{duration: 250}}" role="region" aria-label="Current shown cluster" >
            <span class="text-center">Cluster {shownCluster.cluster_id} Shards:</span>
            <div class="flex flex-row flex-wrap gap-2 p-4 justify-center">
                {#each shownCluster.shards as shard}
                <div class="shard aspect-square p-2 tooltip indicator {shard.status}">
                    {#if shard.shard_id == shownShardID && findCluster} <span class="indicator-item status status-info status-lg"></span> {/if}
                    {shard.shard_id}
                    <div class="tooltip-content flex flex-col">
                        <span>up: {shard.up}</span>
                        <span>latency: {shard.latency}</span>
                        <span>last heartbeat: {dateAgo(shard.last_heartbeat)}</span>
                        <span>last connection: {dateAgo(shard.last_connection)}</span>
                    </div>
                </div>
                {/each}
            </div>
        </div>
        {/if}

        <div class="divider"></div>
        
        <div role="region" aria-label="Cluster/Shard locator" class="flex flex-col">
            <h2 class="text-lg">Find My Shard/Cluster:</h2>
            <span class="text-sm pb-4">Enter a server ID or a message link to find the shard currently assigned to your server.</span>
            <input type="text" aria-label="Server ID or Message Link Input" placeholder="Server ID or Message Link" class="input {findClusterErr != "" ? "input-error" : ""}" bind:value={findClusterInput} on:input={clusterInfoHandler} />
            {#if findClusterErr != ""}
                <span class="text-sm text-error">{findClusterErr}</span>
            {/if}
            {#if findClusterInput != "" && findClusterErr == "" && showCluster}
                <span class="text-md text-info pt-4">You are on cluster {shownCluster.cluster_id}, shard {shownShardID}!</span>
            {/if}
        </div>
    </div>
</div>

<style>
    :root {
        --cluster-item-size: 4.2rem;
        --cluster-gap-size: calc(2*var(--spacing));
    }
    .cluster
    {
        width: var(--cluster-item-size);
        height: var(--cluster-item-size);
        border-radius: 4px;
        background-color: #888888;
        justify-content: center;
        align-items: center;
        cursor: pointer;
    }
    .cluster-ctr 
    {
        /* there's probably a better way to do this that i'm not seeing */
        /* i spent wayyyyy too long on this tho *sobs* */
        margin-left: calc(0.5 * mod(100%, var(--cluster-item-size) + var(--cluster-gap-size)));
        gap: var(--cluster-gap-size);
    }
    .shard
    {
        width: 3rem;
        height: 3rem;
        border-radius: 4px;
        background-color: #888888;
        justify-content: center;
        align-items: center;
    }
    .healthy {
        background-color: #00cc00;
    }
    .degraded {
        background-color: #da9317;
    }
    .severe {
        background-color: #cc0000;
    }
</style>