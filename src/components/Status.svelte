<script lang="ts">
    import { onMount } from 'svelte';
    import { SvelteMap, SvelteSet } from 'svelte/reactivity';
    import { slide } from 'svelte/transition';
    import { type Incident, type Status } from '$lib/types.ts';
    import { api_url, dateAgo } from '$lib/util';

    function toggleIncidentDetails(id: string){
        shownIncidentDetails.set(id, !shownIncidentDetails.get(id));
    }

    let incidents: SvelteMap<string, Incident> = $state(new SvelteMap());
    let active_incidents: Incident[] = $state([]);

    let shownIncidentDetails: SvelteMap<string, boolean> = $state(new SvelteMap<string, boolean>());
    let status: Status | undefined = $state();
    let error: any = $state();

    onMount(async () => {
        try {
            const response = await fetch(api_url + "/status")
            const data = await response.json();
            status = data;
        } catch (e) {
            error = e;
            console.error(e);
        }

        if (status && status.active_incidents.length > 0) {
            try {
                const response = await fetch(api_url + "/incidents")
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
                    shownIncidentDetails.set(id, false);
                    let incident = incidents.get(id);
                    if(incident) active_incidents.push(incident);
                })

                active_incidents.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime())
            } catch (e) {
                error = e;
                console.error(e);
            }
        }
    })
</script>

<div class="card p-4">

<div role="alert" class="alert alert-error">
    <span class="text-center text-lg">Major System Outage!</span>
</div>

{#if status && status.active_incidents.length > 0 && !error}
    <div class="w-full flex flex-col gap-4 py-4" role="region" aria-label="Active incidents">
        {#each active_incidents as incident}
            {@const impact_class = incident.impact == "none" ? "badge-neutral" : incident?.impact == "minor" ? "badge-warning" : "badge-error"}
            {@const incident_class = incident.updates && incident.updates.length > 0 ? "incident" : ""}
            <button class="card bg-base-200 w-full shadow-sm {incident_class}" onclick={()=>{toggleIncidentDetails(incident.id)}}>
                <div class="card-body">
                    <h2 class="card-title">
                        <div class="flex flex-row w-full">
                            {incident.name}
                            <div class="ml-auto items-end">
                                <div class="badge {impact_class}">{incident.impact}</div>
                            </div>
                        </div>
                    </h2>
                    <span class="text-left">{incident.description}</span>
                    {#if shownIncidentDetails.get(incident.id) && incident.updates && incident.updates.length > 0}
                        <div transition:slide="{{duration: 250}}">
                            <div class="divider"></div>
                            <ul class="timeline timeline-vertical timeline-compact gap-4">
                                {#each incident.updates as update}
                                <li>
                                    <div class="timeline-start">{dateAgo(update.timestamp)}</div>
                                    <hr />
                                    <div class="timeline-end timeline-box">{update.text}</div>
                                </li>
                                {/each}
                            </ul>
                        </div>
                    {/if}
                    {#if incident.timestamp}
                        <div class="card-actions justify-end">Started {dateAgo(incident.timestamp)}</div>
                    {/if}
                </div>
            </button>
        {/each}
    </div>
{/if}
</div>

<style>
    .incident {
        cursor: pointer;
    }
</style>