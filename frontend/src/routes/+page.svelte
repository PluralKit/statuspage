<script lang="ts">
    import { themeChange } from 'theme-change';
    import { onMount } from 'svelte';

    import NavbarComponent from '$components/Navbar.svelte';
    import StatusComponent from '$components/Status.svelte';
    import ClustersComponent from '$components/Clusters.svelte';

    import { type Status, type Incident, type ClustersWrapper} from '$lib/types.ts';
    import { SvelteMap } from 'svelte/reactivity';

    let error = $state();
    let status: Status | undefined = $state();
    let incidents: SvelteMap<string, Incident> = $state(new SvelteMap());
    let active_incidents: Incident[] = $state([]);
    let clustersInfo: ClustersWrapper | undefined = $state();

    async function fetchData() {
        try {
            const response = await fetch("/api/v1/status")
            const data = await response.json();
            status = {
                ...data,
                timestamp: new Date(data.timestamp)
            };
        } catch (e) {
            console.error('Error fetching data:', e);
            error = e;
        }

        if (status && status.active_incidents.length > 0) {
            try {
                const response = await fetch("/api/v1/incidents/active");
                const data = await response.json();
                const entries = Object.entries(data.incidents).map(([id, incidentData]: [string, any]) => {
                    const incident: Incident = {
                        ...incidentData,
                        timestamp: new Date(incidentData.timestamp),
                        last_update: new Date(incidentData.last_update),
                        resolution_timestamp: incidentData.resolution_timestamp
                            ? new Date(incidentData.resolution_timestamp)
                            : null,
                        updates: (incidentData.updates || []).map((update: any) => ({
                            ...update,
                            timestamp: new Date(update.timestamp)
                        })),
                    };
                    return [id, incident] as [string, Incident];
                });
                incidents = new SvelteMap<string, Incident>(entries);
                incidents.forEach((i)=>{
                    i.updates?.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());
                });

                status.active_incidents.forEach((id)=>{
                    let incident = incidents.get(id);
                    if(incident) active_incidents.push(incident);
                })

                active_incidents.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime())
            } catch (e) {
                console.error('Error fetching data:', e);
                error = e;
            }
        }

        try {
            const response = await fetch("/api/v1/clusters");
            const data = await response.json() as ClustersWrapper;
            clustersInfo = data
            if (!clustersInfo.clusters) {
                throw new Error("clusters is undefined")
            }
            clustersInfo.clusters.forEach((cluster, i) => {
                if (!cluster.up) cluster.status = "down";
                else if (cluster.shards_up < (data.max_concurrency/2)) cluster.status = "down";
                else if (cluster.avg_latency < 200) cluster.status = "healthy";
                else if (cluster.avg_latency < 400) cluster.status = "degraded";
                else cluster.status = "severe";
                cluster.id = i
            });
        } catch (e) {
            console.error('Error fetching data:', e);
            error = e;
        }
    }

    onMount(()=>{
        themeChange(false);
        fetchData();
    })
</script>

<svelte:head>
	<title>PluralKit Status</title>
    <meta property="og:title" content="PluralKit Status" />
    <meta name="description" content="The status page for PluralKit -- shard and incident information.">
    <meta property="og:description" content="The status page for PluralKit -- shard and incident information.">
    <meta property="og:url" content="https://status.pluralkit.me">
    <meta name="theme-color" content="#DA9317">
    <meta property="og:type" content="website" />
    <meta property="og:image" content="https://status.pluralkit.me/myriad_write.webp" />
</svelte:head>

<div class="w-full justify-center p-4">
    <div class="w-full 2xl:w-1/2 lg:w-3/4 m-auto flex-col">
        <NavbarComponent />
        <div class="flex flex-col gap-4">
            <StatusComponent incidents={active_incidents} status={status} error={error}/>
            <ClustersComponent clustersInfo={clustersInfo} error={error}/>
        </div>
    </div>
</div>