import { error, type HttpError } from '@sveltejs/kit';
import { type Incident } from '$lib/types.ts';
import { env } from '$env/dynamic/private'

export async function load({ fetch, params }) {
    const id = params.id;
    if (id.length != 8) throw error(400, 'Invalid ID');
    let incident: Incident | undefined = undefined;
    try {
        const response = await fetch(`${env.BACKEND_URL}/api/v1/incidents/${id}`);
        if(response.status == 404) throw error(404, 'Incident not found');
        if(!response.ok) throw error(500, 'Error while loading incident');
        const data = await response.json();
        
        incident = {
            ...data,
            timestamp: new Date(data.timestamp),
            last_update: new Date(data.last_update),
            resolution_timestamp: data.resolution_timestamp
                ? new Date(data.resolution_timestamp)
                : null,
            updates: (data.updates || []).map((update: any) => ({
                ...update,
                timestamp: new Date(update.timestamp)
            })),
        };
    } catch (e: any) {
        if(e.status == 404) throw error(404, 'Incident not found');
        console.error('Error fetching data:', e);
        throw error(500, 'Error while loading incident');
    }
    
    return { incident };
}
