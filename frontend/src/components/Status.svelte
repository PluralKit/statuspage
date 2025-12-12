<script lang="ts">
    import IncidentComponent from './Incident.svelte';

    let { incidents, status, error } = $props();

    let statusText: string | undefined = $state();
    let statusInfoText: string | undefined = $state();
    let statusClass: string | undefined = $state();
    
    $effect(() => {
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
    });    
</script>

<div class="card">
    <div role="alert" class="alert {statusClass} flex flex-col items-start gap-1 h-1/8">
        {#if status}
        <span class="text-lg font-bold">{statusText}</span>
        <span class="text-md">{statusInfoText}</span>
        <span class="text-xs italic pt-2">Last refreshed status at {status.timestamp.toLocaleTimeString()}</span>
        {/if}
    </div>

    {#if status && status.active_incidents.length > 0 && !error}
        <div class="w-full flex flex-col gap-4 py-4" role="region" aria-label="Active incidents">
            {#each incidents as incident}
                <IncidentComponent incident={incident} />
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