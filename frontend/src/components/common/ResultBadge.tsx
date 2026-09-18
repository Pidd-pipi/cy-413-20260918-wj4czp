import {Tag} from 'antd';export function ResultBadge({result}:{result:string}){return <Tag color={result.includes('关照')?'orange':'green'}>{result}</Tag>}
