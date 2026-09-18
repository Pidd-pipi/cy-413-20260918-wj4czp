import type {Mood} from '../types'; let moods:Mood[]=[];export const moodStore={get:()=>moods,set:(values:Mood[])=>{moods=values}};
