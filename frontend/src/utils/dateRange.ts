export function lastDays(days=7){return Array.from({length:days},(_,i)=>{const d=new Date();d.setDate(d.getDate()-(days-i-1));return d.toISOString().slice(0,10)})}
