<script lang="ts">
    import { dateAgo, discordTimestamp } from '$lib/util';
    import type { Incident } from '$src/lib/types';
    import { marked } from 'marked';
    import { Render } from 'svelte-purify/browser-only';

    export let incident: Incident, full = false;
    let impact_class = incident?.impact == "none" ? "badge-neutral" : incident?.impact == "minor" ? "badge-warning" : "badge-error";
    let incident_class = incident?.updates && incident?.updates.length > 0 ? "incident" : "";

    marked.use({ extensions: [discordTimestamp] });
</script>
{#if incident != undefined}
{#if full}
<a class="card bg-base-200 w-full shadow-sm {incident_class}" href={!full ? `/i/${incident.id}` : undefined}>
    <div class="card-body w-full items-center">
        <h2 class="card-title py-4">
            <div class="flex flex-col w-full items-center">
                <span class="text-2xl">{incident.name} {#if incident.impact != "none"}<div class="badge {impact_class}">{incident.impact}</div>{/if}</span>
                <span class="text-sm italic">{dateAgo(incident.timestamp.getTime())}</span>
            </div>
        </h2>
        <span class=" flex flex-col gap-4 text-left text-base">
            {#await marked(incident.description)}
                <p>loading...</p>
            {:then html}
                <Render html={html} />
            {/await}
        </span>
        {#if incident.updates && incident.updates.length > 0}
            <div>
                <div class="divider"></div>
                <ul class="timeline timeline-vertical timeline-compact gap-6 py-2">
                    {#each incident.updates.sort((a, b) => a.timestamp.getTime() + b.timestamp.getTime()) as update}
                    <li>
                        <div class="timeline-start timeline-box flex gap-4 py-3 text-left text-sm">
                            {#if update.status}
                                <span class="update-status font-bold">{String(update.status).charAt(0).toUpperCase() + String(update.status).slice(1)}</span>
                            {/if}
                            {#await marked(update.text)}
                                <p>loading...</p>
                            {:then html}
                                <Render html={html} />
                            {/await}
                        </div>
                        <div class="timeline-end text-xs pl-4">
                            {dateAgo(update.timestamp.getTime())}
                        </div>
                    </li>
                    {/each}
                </ul>
            </div>
        {/if}
    </div>
</a>
{:else}
<a class="card bg-base-200 w-full shadow-sm {incident_class}" href={!full ? `/i/${incident.id}` : undefined}>
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
        <div class="card-actions pt-2">
            <div class="justify-end ml-auto">
                <span class="text-sm italic">{dateAgo(incident.timestamp.getTime())} | {incident.id}</span>
            </div>
        </div>
        {#if incident.updates && incident.updates.length > 0}
            {@const update = incident.updates.sort((a, b) => a.timestamp.getTime() + b.timestamp.getTime())[0]}
            <div>
                <div class="divider"></div>
                <ul class="timeline timeline-vertical timeline-compact gap-4">
                    <li>
                        <div class="timeline-start timeline-box flex gap-4 py-3 text-left text-sm">
                            {#if update.status}
                                <span class="update-status font-bold">{String(update.status).charAt(0).toUpperCase() + String(update.status).slice(1)}</span>
                            {/if}
                            {#await marked(update.text)}
                                <p>loading...</p>
                            {:then html}
                                <Render html={html} />
                            {/await}
                        </div>
                        <div class="timeline-end text-xs pl-4">
                            {dateAgo(update.timestamp.getTime())}
                        </div>
                    </li>
                </ul>
            </div>
        {/if}
    </div>
</a>
{/if}
{/if}

<style>
    .update-text {
        line-break: auto;
    }
</style>