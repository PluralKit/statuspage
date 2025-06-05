<script lang="ts">
    import { themeChange } from 'theme-change';
    import { onMount } from 'svelte';

    import Navbar from '$components/Navbar.svelte';
    import Status from '$components/Status.svelte';
    import Clusters from '$components/Clusters.svelte';

    onMount(()=>{
        themeChange(false);
    })
</script>

<svelte:head>
	<title>PluralKit Status</title>
</svelte:head>

<div class="w-full justify-center p-4">
    <div class="w-full 2xl:w-1/2 lg:w-3/4 m-auto flex-col">
        <Navbar />

        <div class="flex flex-col gap-4">
            <Status />
            <Clusters />
        </div>
    </div>
</div>