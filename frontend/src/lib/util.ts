// @ts-nocheck
//kinda janky but it works
export function dateAgo(date: number) {
  if (date == 0) return "never";
  // difference in milliseconds
  const msDifference = date - Date.now();
  // convert to seconds
  const diffSeconds = msDifference / 1000;
  const diffMinutes = diffSeconds / 60;
  const diffHours = diffMinutes / 60;
  const diffDays = diffHours / 24;

  if(Math.abs(diffDays) > 1) return new Date(date).toLocaleString('en-us', { dateStyle: 'short', timeStyle: 'short' });

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

const timestampRegex = /^<t:(\d+)(?::([tTdDfFR]))?>/;
export const discordTimestamp = {
  name: 'discordTimestamp',
  level: 'inline',
  start(src) {
    return src.match(/<t:/)?.index;
  },
  tokenizer(src, tokens) {
    const match = timestampRegex.exec(src);
    if (match) {
      const [raw, timestamp, format] = match;

      return {
        type: 'discordTimestamp',
        raw: raw,
        timestamp: timestamp,
        format: format || 'f',
      };
    }
  },
  renderer(token) {
    const date = new Date(parseInt(token.timestamp, 10) * 1000);
    const isoString = date.toISOString();
    let text = '';

    switch (token.format) {
      case 't':
        text = date.toLocaleTimeString();
        break;
      case 'T':
        text = date.toLocaleTimeString('en-US', { timeStyle: 'medium' });
        break;
      case 'd':
        text = date.toLocaleDateString();
        break;
      case 'D':
        text = date.toLocaleDateString('en-US', { dateStyle: 'long' });
        break;
      case 'F':
        text = date.toLocaleString('en-US', { dateStyle: 'long', timeStyle: 'short' });
        break;
      case 'R':
      case 'f':
      default:
        text = date.toLocaleString();
        break;
    }
    return `<span class="bg-base-300 p-1 rounded" title="${date.toString()}"><time datetime="${isoString}">${text}</time></span>`;
  },
};