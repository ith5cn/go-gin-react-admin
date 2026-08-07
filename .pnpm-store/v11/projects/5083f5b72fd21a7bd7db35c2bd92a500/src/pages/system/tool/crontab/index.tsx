import { crontabDeleteApi, crontabListApi, crontabRunApi } from "@/api/system/crontab";
import Ith5Select from "@/components/ith5ui/ith5-select";
import Ith5Table, { type TableRef } from "@/components/ith5ui/ith5-table";
import { Button, Col, Form, Input, message } from "antd";
import { FileTextOutlined, PlayCircleOutlined } from "@ant-design/icons";
import { useRef } from "react";
import CrontabEdit, { type CrontabEditRef } from "./edit";
import CrontabLog, { type CrontabLogRef } from "./log";

type CrontabRecord = Record<string, unknown> & { id: number | string };

export default function Crontab() {
    const tableRef = useRef<TableRef>(null);
    const editRef = useRef<CrontabEditRef>(null);
    const logRef = useRef<CrontabLogRef>(null);

    const options = {
        api: crontabListApi,
        add: {
            show: true,
            auth: ["system/crontab/create"],
            func: () => editRef.current?.open("add"),
        },
        edit: {
            show: true,
            auth: ["system/crontab/update"],
            func: (record: CrontabRecord) => editRef.current?.open("edit", record),
        },
        delete: {
            show: true,
            auth: ["system/crontab/destroy"],
            func: async (record: CrontabRecord) => {
                const res = await crontabDeleteApi(record.id);
                if (res.code === 0) {
                    message.success('删除成功');
                    tableRef.current?.refresh();
                }
            }
        }
    }

    return (
        <>
            <Ith5Table
                ref={tableRef}
                options={options}
                operationBeforeExtend={(record) => (
                    <>
                    <Button 
                        type="link" 
                        size="small" 
                        icon={<PlayCircleOutlined />} 
                        onClick={async () => {
                            const res = await crontabRunApi(record.id);
                            if (res.code === 0) {
                                message.success(res.data || '执行已触发');
                            }
                        }}
                    >
                        执行一次
                    </Button>
                    <Button type="link" size="small" icon={<FileTextOutlined />} onClick={async()=>{
                        logRef.current?.open({id:record.id});
                    }}>日志</Button>
                    </>
                )}
                columns={[
                    {
                        title: '任务名称',
                        dataIndex: 'name',
                        key: 'name',
                    },
                    {
                        title: '任务类型',
                        dataIndex: 'type',
                        key: 'type',
                        type: 'dict',
                        dict: 'crontab_type'
                    },
                    {
                        title: '执行类型',
                        dataIndex: 'taskStyle',
                        key: 'taskStyle',
                        render: (value: number | string) => Number(value) === 2 ? 'HTTP 请求任务' : '系统内部任务',
                    },
                    {
                        title: '调用目标',
                        dataIndex: 'target',
                        key: 'target',
                    },
                    {
                        title: 'Cron表达式',
                        dataIndex: 'rule',
                        key: 'rule',
                    },
                    {
                        title: '任务状态',
                        dataIndex: 'status',
                        key: 'status',
                        type: 'dict',
                        dict: 'status'
                    },
                    {
                        title: '备注',
                        dataIndex: 'remark',
                        key: 'remark',
                    },
                ]}
            >
                <Col span={6}>
                    <Form.Item name="name" label="任务名称">
                        <Input placeholder="请输入任务名称" allowClear />
                    </Form.Item>
                </Col>
                <Col span={6}>
                    <Form.Item name="status" label="状态">
                        <Ith5Select dict="status" placeholder="请选择状态" />
                    </Form.Item>
                </Col>
            </Ith5Table>
            <CrontabEdit ref={editRef} onSuccess={() => tableRef.current?.refresh()} />
            <CrontabLog ref={logRef}/>
        </>
    );
}