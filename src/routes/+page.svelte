<script lang="ts">
    import { onMount } from 'svelte';
    import { slide } from 'svelte/transition';
    import { themeChange } from 'theme-change';

    const stats_url = "https://api.pluralkit.me/private/discord/shard_state"

    //uhhhh don't @me, i wrote like 90% of this at 11pm while not awake, todo: cleanup

    interface Shard {
        shard_id: number;
        cluster_id: number;
        up: boolean;
        status: string;
        latency: number;
        last_heartbeat: Date;
        last_connection: Date;
    }

    interface Cluster {
        cluster_id: number;
        avg_latency: number;
        up: boolean;
        status: string;
        shards: Shard[];
    }

    let clusters: Cluster[] = [];
    let shards_up = 0;
    let shards_total = 0;
    let avg_latency = 0;
    let max_concurrency = 16;
    let error;

    onMount(async () => {
        themeChange(false);
        try {
            const response = await fetch(stats_url);
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

    //kinda janky but it works
    function dateAgo(date: Date) {
		// difference in milliseconds
		const msDifference = Math.abs((Number(date) - Date.now()));
		// convert to seconds
		const diffSeconds = msDifference / 1000;
        const diffMinutes = diffSeconds / 60;
        const diffHours = diffMinutes / 60;
        const diffDays = diffHours / 24;

        let diff = "";
        if(diffDays > 1) diff = Math.floor(diffDays) + " days ago";
        else if (diffHours > 1) diff = Math.floor(diffHours) + " hours ago";
        else if (diffMinutes > 1) diff = Math.floor(diffMinutes) + " minutes ago";
        else diff = Math.floor(diffSeconds) + " seconds ago";

        return diff;
	}

    const getShardID = (guild_id: string) => guild_id == "" ? -1 : Number((BigInt(guild_id) >> BigInt(22)) % BigInt(shards_total));

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
            let shardID = getShardID(match[1]);
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
            var shardID = getShardID(findClusterInput);
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

    //kinda a janky fix for closing, but whatevs it'll work for now
    function clickHandler(event: MouseEvent){
        const clicked = event.target as HTMLElement;
        if(clicked === null || ! clicked.tagName) return;
        if(clicked.tagName !== 'BUTTON' && clicked.tagName !== 'INPUT' && clicked.tagName !== 'A' 
            && !clicked.classList.contains("cluster") && !clicked.classList.contains("shard") && !clicked.classList.contains("btn")){
            if(!findCluster) showCluster = false;
        }
    }
</script>

<svelte:head>
	<title>PluralKit Status</title>
</svelte:head>

<svelte:body onclick={clickHandler} />

<div class="w-full justify-center p-4">
    <div class="w-full 2xl:w-1/2 m-auto flex-col">
        <div class="navbar pb-6" role="region" aria-label="Navigation menu">
            <div class="flex-1 pl-2 navbar-start">
                <div class="avatar pr-4">
                    <div class="w-12 rounded">
                      <img alt="An icon of myriad looking down and writing on a notepad" src="myriad_write.png" />
                    </div>
                </div>
                <h1 class="text-2xl">PluralKit Status</h1>
            </div>

            <div class="navbar-end">
                <div class="dropdown dropdown-end">
                    <div tabindex="0" role="button" class="btn btn-ghost lg:hidden">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"> <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h8m-8 6h16" /> </svg>
                    </div>
                    <ul class="menu menu-md dropdown-content mt-3 z-[1] p-2 shadow bg-base-100 rounded-box w-52">
                        <li><a href="https://discordstatus.com/">Discord Status</a></li>
                        <li><a href="https://stats.pluralkit.me">Statistics</a></li>
                        <li><a href="https://discord.gg/PczBt78">Support Server</a></li>
                    </ul>
                </div>
                <ul class="menu menu-horizontal px-1 hidden lg:flex">
                    <li><a href="https://discordstatus.com/">Discord Status</a></li>
                    <li><a href="https://stats.pluralkit.me">Statistics</a></li>
                    <li><a href="https://discord.gg/PczBt78">Support Server</a></li>
                     <li>
                        <label class="swap swap-rotate" aria-label="Dark/Light Mode Toggle">
                            <input type="checkbox" class="theme-controller" data-toggle-theme="dark,light"/>

                            <svg class="swap-off h-5 w-5 fill-current" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M5.64,17l-.71.71a1,1,0,0,0,0,1.41,1,1,0,0,0,1.41,0l.71-.71A1,1,0,0,0,5.64,17ZM5,12a1,1,0,0,0-1-1H3a1,1,0,0,0,0,2H4A1,1,0,0,0,5,12Zm7-7a1,1,0,0,0,1-1V3a1,1,0,0,0-2,0V4A1,1,0,0,0,12,5ZM5.64,7.05a1,1,0,0,0,.7.29,1,1,0,0,0,.71-.29,1,1,0,0,0,0-1.41l-.71-.71A1,1,0,0,0,4.93,6.34Zm12,.29a1,1,0,0,0,.7-.29l.71-.71a1,1,0,1,0-1.41-1.41L17,5.64a1,1,0,0,0,0,1.41A1,1,0,0,0,17.66,7.34ZM21,11H20a1,1,0,0,0,0,2h1a1,1,0,0,0,0-2Zm-9,8a1,1,0,0,0-1,1v1a1,1,0,0,0,2,0V20A1,1,0,0,0,12,19ZM18.36,17A1,1,0,0,0,17,18.36l.71.71a1,1,0,0,0,1.41,0,1,1,0,0,0,0-1.41ZM12,6.5A5.5,5.5,0,1,0,17.5,12,5.51,5.51,0,0,0,12,6.5Zm0,9A3.5,3.5,0,1,1,15.5,12,3.5,3.5,0,0,1,12,15.5Z" /></svg>

                            <svg class="swap-on h-5 w-5 fill-current" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M21.64,13a1,1,0,0,0-1.05-.14,8.05,8.05,0,0,1-3.37.73A8.15,8.15,0,0,1,9.08,5.49a8.59,8.59,0,0,1,.25-2A1,1,0,0,0,8,2.36,10.14,10.14,0,1,0,22,14.05,1,1,0,0,0,21.64,13Zm-9.5,6.69A8.14,8.14,0,0,1,7.08,5.22v.27A10.15,10.15,0,0,0,17.22,15.63a9.79,9.79,0,0,0,2.1-.22A8.11,8.11,0,0,1,12.14,19.73Z" /></svg>
                        </label>
                    </li>
                </ul>
            </div>
        </div>
        <div class="card bg-base-200 shadow-sm">
            <div class="p-8 flex flex-col">
                <div class="stats shadow" role="region" aria-label="Overall statistics">
                    <div class="stat">
                      <div class="stat-title">Shards Up</div>
                      <div class="stat-value">{shards_up} / {shards_total}</div>
                    </div>
                    <div class="stat">
                        <div class="stat-title"> Average Latency</div>
                        <div class="stat-value">{avg_latency} ms</div>
                    </div>
                </div>

                <div class="divider"></div>

                <div class="w-full flex flex-col" role="region" aria-label="Cluster status">
                    <h2 class="text-lg">Clusters:</h2>
                    <div class="flex flex-wrap flex-row gap-2 py-6 justify-center">
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
                <div class="card bg-base-300 p-8" transition:slide="{{duration: 250}}" role="region" aria-label="Current shown cluster" >
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
    </div>
</div>

<style>
.cluster
{
    width: 4em;
    height: 4em;
    border-radius: 4px;
    background-color: #888888;
    justify-content: center;
    align-items: center;
    cursor: pointer;
}
.shard
{
    width: 3em;
    height: 3em;
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