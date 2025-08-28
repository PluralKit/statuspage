//kinda janky but it works
export function dateAgo(date: number) {
  if (date == 0) return "never";
  // difference in milliseconds
  const msDifference = (date * 1000) - Date.now();
  // convert to seconds
  const diffSeconds = msDifference / 1000;
  const diffMinutes = diffSeconds / 60;
  const diffHours = diffMinutes / 60;
  const diffDays = diffHours / 24;

  

  const rtf = new Intl.RelativeTimeFormat("en", { numeric: "auto" });

  let diff = "";
  if (Math.abs(diffDays) > 1) diff = rtf.format(Math.floor(diffDays), "day");
  else if (Math.abs(diffHours) > 1)
    diff = rtf.format(Math.floor(diffHours), "hour");
  else if (Math.abs(diffMinutes) > 1)
    diff = rtf.format(Math.floor(diffMinutes), "minute");
  else diff = rtf.format(Math.floor(diffSeconds), "second");

  return diff;
}

export const getShardID = (guild_id: string, shards_total: number) =>
  guild_id == ""
    ? -1
    : Number((BigInt(guild_id) >> BigInt(22)) % BigInt(shards_total));
