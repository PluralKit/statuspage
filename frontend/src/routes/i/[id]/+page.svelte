<script lang="ts">
    import { themeChange } from 'theme-change';
    import { onMount } from 'svelte';

    import Navbar from '$components/Navbar.svelte';
    import Incident from '$components/Incident.svelte';

    export let data;

    const og_title = data.incident ? `Incident - ${data.incident.name}` : "PluralKit Status";
    const og_desc = data.incident ? `${data.incident.description}` : "The status page for PluralKit -- shard and incident information.";
    const og_link = data.incident ? `https://status.pluralkit.me/i/${data.incident.id}` : "https://status.pluralkit.me";
    const og_color = data.incident ? 
        (data.incident.impact == "none" ? "#99c1f1" : data.incident.impact == "minor" ? "#fcb700" : "#ff637d")
        : "#DA9317";
    const og_date = data.incident ? data.incident.timestamp.toISOString() : undefined;

    onMount(()=>{
        themeChange(false);
    })
</script>

<svelte:head>
	<title>PluralKit Status</title>
    <meta property="og:site_name" content="PluralKit Status" />
    <meta property="og:title" content={og_title} />
    <meta name="description" content={og_desc} />
    <meta property="og:description" content={og_desc} />
    <meta property="og:url" content={og_link} />
    <meta name="theme-color" content={og_color} />
    <meta property="og:type" content="website" />
</svelte:head>

<div class="w-full justify-center p-4">
    <div class="w-full 2xl:w-1/2 lg:w-3/4 m-auto flex-col">
        <Navbar />
        <div class="flex flex-col gap-4">
            {#if data.incident}
            <Incident incident={data.incident} full={true}/>
            {/if}
        </div>
    </div>
</div>