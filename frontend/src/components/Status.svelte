<script lang="ts">
    import { onMount } from 'svelte';
    import { SvelteMap, SvelteSet } from 'svelte/reactivity';
    import { slide } from 'svelte/transition';
    import { type Incident, type Status } from '$lib/types.ts';
    import { dateAgo } from '$lib/util';

    import { marked } from 'marked';
    import { Render } from 'svelte-purify/browser-only'

    function toggleIncidentDetails(id: string){
        shownIncidentDetails.set(id, !shownIncidentDetails.get(id));
    }

    let incidents: SvelteMap<string, Incident> = $state(new SvelteMap());
    let active_incidents: Incident[] = $state([]);

    let shownIncidentDetails: SvelteMap<string, boolean> = $state(new SvelteMap<string, boolean>());
    
    let status: Status | undefined = $state();
    let statusText: string | undefined = $state();
    let statusInfoText: string | undefined = $state();
    let statusClass: string | undefined = $state();
    
    let error: any = $state();

    async function fetchStatus() {
        try {
            const response = await fetch("/api/v1/status")
            const data = await response.json();
            status = {
                ...data,
                timestamp: new Date(data.timestamp)
            };
        } catch (e) {
            error = e;
            console.error(e);
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

        // TODO: better wording here? i now realize that 'systems' could be confused for literal pk systems
        switch (status?.status) {
            case "operational":
                statusText = "All systems operational!"
                statusInfoText = "There are no active known incidents."
                statusClass = "alert-success"
                break;
            case "degraded":
                statusText = "Some systems degraded!"
                statusInfoText = "Some things might not work properly, see incidents listed below for details."
                statusClass = "alert-warning"
                break;
            case "major_outage":
                statusText = "Major systems outage!"
                statusInfoText = "Most things probably aren't functioning, see incidents listed below for details."
                statusClass = "alert-error"
                break;
            default:
                break;
        }
    }

    onMount(async () => {
        await fetchStatus();
    })
</script>




<div class="card">
    <div role="alert" class="alert {statusClass} flex flex-col items-start gap-1 h-28">
        {#if status}
        <span class="text-lg font-bold">{statusText}</span>
        <span class="text-md">{statusInfoText}</span>
        <span class="text-xs italic pt-2">Last refreshed status at {status.timestamp.toLocaleTimeString()}</span>
        {/if}
    </div>

    {#if status && status.active_incidents.length > 0 && !error}
        <div class="w-full flex flex-col gap-4 py-4" role="region" aria-label="Active incidents">
            {#each active_incidents as incident}
                {@const impact_class = incident.impact == "none" ? "badge-neutral" : incident?.impact == "minor" ? "badge-warning" : "badge-error"}
                {@const incident_class = incident.updates && incident.updates.length > 0 ? "incident" : ""}
                <button class="card bg-base-200 w-full shadow-sm {incident_class}" onclick={()=>{toggleIncidentDetails(incident.id)}}>
                    <div class="card-body w-full">
                        <h2 class="card-title">
                            <div class="flex flex-row w-full">
                                {incident.name}
                                <div class="ml-auto items-end">
                                    {#if incident.impact != "none"}<div class="badge {impact_class}">{incident.impact}</div>{/if}
                                </div>
                            </div>
                        </h2>
                        <span class=" flex flex-col gap-4 text-left text-sm">
                            {#await marked(incident.description)}
                                <p>loading...</p>
                            {:then html}
                                <Render html={html} />
                            {/await}
                        </span>
                        {#if shownIncidentDetails.get(incident.id) && incident.updates && incident.updates.length > 0}
                            <div transition:slide="{{duration: 250}}">
                                <div class="divider"></div>
                                <ul class="timeline timeline-vertical timeline-compact gap-4">
                                    {#each incident.updates as update}
                                    <li>
                                        <div class="timeline-start">{dateAgo(update.timestamp)}</div>
                                        <div class="timeline-end timeline-box flex flex-col gap-4 p-4 text-left text-sm">
                                            {#await marked(update.text)}
                                                <p>loading...</p>
                                            {:then html}
                                                <Render html={html} />
                                            {/await}
                                        </div>
                                    </li>
                                    {/each}
                                </ul>
                            </div>
                        {/if}
                        <div class="card-actions pt-2">
                            {#if incident.updates && incident.updates.length > 0}
                            <div class="justify-start">
                                <span class="text-sm italic justify-start">(click to {shownIncidentDetails.get(incident.id) ? "hide" : "show"} updates)</span>
                            </div>
                            {/if}
                            <div class="justify-end ml-auto">
                                <span class="text-sm italic">Started {dateAgo(incident.timestamp)} | {incident.id}</span>
                            </div>
                        </div>
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
    .update-text {
        
        line-break: auto;
    }
</style>