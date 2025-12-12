<script lang="ts">
    import { slide } from 'svelte/transition';
    import { dateAgo, getShardID } from '$lib/util';
    import { type Cluster, type ClustersWrapper, type Shard } from '$lib/types';

    let {clustersInfo, error}: {clustersInfo?: ClustersWrapper; error: any} = $props();
    
    //kinda a janky fix for closing, but whatevs it'll work for now
    function clickHandler(event: MouseEvent){
        const clicked = event.target as HTMLElement;
        if(clicked === null || ! clicked.tagName) return;
        if(clicked.tagName !== 'BUTTON' && clicked.tagName !== 'INPUT' && clicked.tagName !== 'A' 
            && !clicked.classList.contains("cluster") && !clicked.classList.contains("shard") && !clicked.classList.contains("btn")){
            if(!findCluster) shownCluster = undefined;
        }
    }

    async function getShards(clusterID: number) {
        if (!clustersInfo) return;
        try {
            if (!clustersInfo.clusters) return
            let cluster = clustersInfo.clusters[clusterID]
            const response = await fetch(`/api/v1/clusters/${clusterID}`);
            const data = await response.json() as Shard[];
            cluster.shards = data

            cluster.shards.forEach(shard => {
                if (!shard.up) shard.status = "down"
                else if (shard.latency < 200) shard.status = "healthy"
                else if (shard.latency < 400) shard.status = "degraded"
                else shard.status = "severe"
            });
            clustersInfo.clusters[clusterID] = cluster
        } catch (e) {
            error = e;
            console.error(e);
        }
    }

    let findClusterInput = $state("");
    let findClusterErr = $state("");
    let shownCluster: Cluster | undefined = $state();
    let shownShardID = $state();
    let findCluster = $state(false);

    async function clusterInfoHandler() {
        if (!clustersInfo) return;
        if (findClusterInput.length < 17) return;
        if(findClusterInput == "") {
            findCluster = false;
            findClusterErr = "";
            shownCluster = undefined;
            shownShardID = -1;
            return;
        }

        try {
            var match = findClusterInput.match(/^(?:https:\/\/(?:[\w]*\.)?discord(?:app)?\.com\/channels\/)?(\d+)(?:\/\d+\/\d+)?$/);
            if(match != null) {
                if (!match[1]) throw new Error();
                let shardID = getShardID(match[1], clustersInfo.num_shards);
                let clusterID = Math.floor(shardID / clustersInfo.max_concurrency);
                if (shardID != -1 && clustersInfo.clusters){
                    await getShards(clusterID)
                    shownShardID = Number(shardID);
                    shownCluster = clustersInfo.clusters[clusterID];
                    findCluster = true;
                    findClusterErr = "";
                    return;
                }
            }
        } catch(e) {
            shownCluster = undefined;
            findCluster = false;
            findClusterErr = "Invalid server ID";
        }
    }

    async function showClusterHandler(id: number) {
        if (!clustersInfo) return;
        if(shownCluster && id === shownCluster.id) {
            shownCluster = undefined;
        } else if (clustersInfo.clusters){
            await getShards(id)
            findClusterInput = "";
            findCluster = false;
            shownCluster = clustersInfo.clusters[id];
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
               <div class="stat-value">{#if clustersInfo}{clustersInfo?.shards_up} / {clustersInfo?.num_shards}{/if}</div> 
            </div>
            <div class="stat">
                <div class="stat-title"> Average Latency</div>
                <div class="stat-value">{#if clustersInfo}{clustersInfo?.avg_latency} ms{/if}</div>
            </div>
        </div>
        <div class="flex flex-col items-center w-full" role="region" aria-label="Cluster status">
            <div class="cluster-ctr flex flex-wrap flex-row py-6 justify-start">
                {#if clustersInfo?.clusters}
                {#each clustersInfo.clusters as cluster}
                <button class="cluster aspect-square tooltip indicator {cluster.status}" onclick={()=>{showClusterHandler(cluster.id)}}>
                    {#if cluster.shards_up < clustersInfo.max_concurrency}
                        <span class="indicator-item status status-error"></span>
                    {/if}
                    {cluster.id}
                    <div class="tooltip-content">
                        avg latency: {cluster.avg_latency}
                    </div>
                </button>
                {/each}
                {/if}
            </div>
        </div>

        {#if shownCluster}
        <div class="card bg-base-100 py-8 px-2" transition:slide="{{duration: 250}}" role="region" aria-label="Current shown cluster" >
            <span class="text-center">Cluster {shownCluster.id} Shards:</span>
            <div class="flex flex-row flex-wrap gap-2 p-4 justify-center">
                {#each shownCluster?.shards || [] as shard}
                <div class="shard aspect-square p-2 tooltip indicator {shard.status}">
                    {#if shard.shard_id == shownShardID && findCluster} <span class="indicator-item status status-info status-lg"></span> {/if}
                    {shard.shard_id}
                    <div class="tooltip-content flex flex-col">
                        <span>up: {shard.up}</span>
                        <span>latency: {shard.latency}</span>
                        <span>last connection: {dateAgo(shard.last_connection * 1000)}</span>
                        <span>last heartbeat: {dateAgo(shard.last_heartbeat * 1000)}</span>
                        <span>last reconnect: {dateAgo(shard.last_reconnect * 1000)}</span>
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
            <input type="text" aria-label="Server ID or Message Link Input" placeholder="Server ID or Message Link" class="input {findClusterErr != "" ? "input-error" : ""}" bind:value={findClusterInput} oninput={clusterInfoHandler} />
            {#if findClusterErr != ""}
                <span class="text-sm text-error">{findClusterErr}</span>
            {/if}
            {#if findClusterInput != "" && findClusterErr == "" && shownCluster}
                <span class="text-md text-info pt-4">You are on cluster {shownCluster.id}, shard {shownShardID}!</span>
            {/if}
        </div>
    </div>
</div>

<style>
    :root {
        --cluster-item-size: 4.2rem;
        --cluster-gap-size: calc(2*var(--spacing));
    }
    @media (max-width: 768px) {
        :root {
            --cluster-item-size: 3.8rem;
        }
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