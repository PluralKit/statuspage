export const stats_url = "https://api.pluralkit.me/private/discord/shard_state"
export const api_url = "http://skye-desktop.den.vixen.lgbt:8080/api/v1"

//kinda janky but it works
export function dateAgo(date: Date) {
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

export const getShardID = (guild_id: string, shards_total: number) => guild_id == "" ? -1 : Number((BigInt(guild_id) >> BigInt(22)) % BigInt(shards_total));
