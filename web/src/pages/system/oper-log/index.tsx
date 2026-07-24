import { useRef, useState } from "react";
import { Button, Col, DatePicker, Descriptions, Form, Input, InputNumber, Modal, Select, Tag, Typography, message } from "antd";
import dayjs, { type Dayjs } from "dayjs";
import Ith5Table, { type ColumnDef, type TableRef } from "@/components/ith5ui/ith5-table";
import { operLogDeleteApi, operLogListApi } from "@/api/system/oper-log";

interface OperLogRecord {
  id: number;
  app?: string;
  method?: string;
  requestData?: string;
  statusCode: number;
  durationMs: number;
  username?: string;
  serviceName?: string;
  router?: string;
  ip?: string;
  ipLocation?: string;
  createTime?: string;
}

const methodOptions = ["POST", "PUT", "PATCH", "DELETE"].map((value) => ({ label: value, value }));
const methodColors: Record<string, string> = {
  POST: "green",
  PUT: "blue",
  PATCH: "cyan",
  DELETE: "red",
};

const normalizeSearchParams = (params: Record<string, unknown>) => {
  const next = { ...params };
  const createTime = next.createTime;
  if (Array.isArray(createTime) && createTime.length === 2) {
    next.createTime = (createTime as Dayjs[]).map((item) => item.format("YYYY-MM-DD HH:mm:ss")).join(",");
  }
  return next;
};

const formatDateTime = (value?: string) => value ? dayjs(value).format("YYYY-MM-DD HH:mm:ss") : "-";

const formatRequestData = (value?: string) => {
  if (!value) return "无请求参数";
  try {
    return JSON.stringify(JSON.parse(value), null, 2);
  } catch {
    return value;
  }
};

const renderStatus = (statusCode: number) => {
  if (!statusCode) return <Tag>历史记录</Tag>;
  if (statusCode >= 500) return <Tag color="error">{statusCode}</Tag>;
  if (statusCode >= 400) return <Tag color="warning">{statusCode}</Tag>;
  if (statusCode >= 300) return <Tag color="processing">{statusCode}</Tag>;
  return <Tag color="success">{statusCode}</Tag>;
};

const OperLogIndex = () => {
  const tableRef = useRef<TableRef>(null);
  const [detail, setDetail] = useState<OperLogRecord>();

  return (
    <>
      <Ith5Table
        ref={tableRef}
        searchFields={
          <>
            <Col span={6}>
              <Form.Item name="username" label="操作用户">
                <Input placeholder="请输入操作用户" allowClear />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="method" label="请求方式">
                <Select placeholder="请选择请求方式" options={methodOptions} allowClear />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="serviceName" label="业务名称">
                <Input placeholder="请输入业务名称" allowClear />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="router" label="请求路由">
                <Input placeholder="请输入请求路由" allowClear />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="ip" label="操作 IP">
                <Input placeholder="请输入操作 IP" allowClear />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="statusCode" label="状态码">
                <InputNumber min={100} max={599} placeholder="HTTP 状态码" style={{ width: "100%" }} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="createTime" label="操作时间">
                <DatePicker.RangePicker showTime style={{ width: "100%" }} />
              </Form.Item>
            </Col>
          </>
        }
        options={{
          api: (params) => operLogListApi(normalizeSearchParams(params)),
          delete: {
            show: true,
            auth: ["system/oper-log/destroy"],
            confirmText: "确定要删除这条操作日志吗？",
            func: async (record: OperLogRecord) => {
              await operLogDeleteApi(record.id);
              message.success("删除成功");
              tableRef.current?.refresh();
            },
          },
        }}
        columns={[
          { title: "操作用户", dataIndex: "username", width: 120, render: (value?: string) => value || "-" },
          {
            title: "请求方式",
            dataIndex: "method",
            width: 100,
            render: (value?: string) => value ? <Tag color={methodColors[value]}>{value}</Tag> : "-",
          },
          { title: "业务名称", dataIndex: "serviceName", width: 150, render: (value?: string) => value || "-" },
          { title: "请求路由", dataIndex: "router", width: 240, render: (value?: string) => value || "-" },
          { title: "结果", dataIndex: "statusCode", width: 100, render: renderStatus },
          { title: "耗时", dataIndex: "durationMs", width: 100, render: (value?: number) => value == null ? "-" : `${value} ms` },
          { title: "操作 IP", dataIndex: "ip", width: 140, render: (value?: string) => value || "-" },
          { title: "操作地点", dataIndex: "ipLocation", width: 120, render: (value?: string) => value || "-" },
          {
            title: "请求参数",
            dataIndex: "requestData",
            width: 100,
            render: (_value: string, record: OperLogRecord) => <Button type="link" onClick={() => setDetail(record)}>查看</Button>,
          },
          { title: "操作时间", dataIndex: "createTime", width: 180, render: formatDateTime },
        ] as ColumnDef[]}
      />

      <Modal open={Boolean(detail)} title="操作日志详情" width={760} footer={null} onCancel={() => setDetail(undefined)}>
        <Descriptions bordered size="small" column={2}>
          <Descriptions.Item label="操作用户">{detail?.username || "-"}</Descriptions.Item>
          <Descriptions.Item label="应用">{detail?.app || "-"}</Descriptions.Item>
          <Descriptions.Item label="请求方式">{detail?.method || "-"}</Descriptions.Item>
          <Descriptions.Item label="HTTP 状态">{detail ? renderStatus(detail.statusCode) : "-"}</Descriptions.Item>
          <Descriptions.Item label="业务名称">{detail?.serviceName || "-"}</Descriptions.Item>
          <Descriptions.Item label="请求耗时">{detail ? `${detail.durationMs ?? 0} ms` : "-"}</Descriptions.Item>
          <Descriptions.Item label="请求路由" span={2}>{detail?.router || "-"}</Descriptions.Item>
          <Descriptions.Item label="操作 IP">{detail?.ip || "-"}</Descriptions.Item>
          <Descriptions.Item label="操作地点">{detail?.ipLocation || "-"}</Descriptions.Item>
          <Descriptions.Item label="操作时间" span={2}>{formatDateTime(detail?.createTime)}</Descriptions.Item>
        </Descriptions>
        <Typography.Title level={5} style={{ marginTop: 20 }}>请求参数</Typography.Title>
        <Typography.Paragraph
          copyable={Boolean(detail?.requestData)}
          style={{ maxHeight: 320, overflow: "auto", whiteSpace: "pre-wrap", wordBreak: "break-all", padding: 12, background: "#f6f8fa", borderRadius: 6 }}
        >
          {formatRequestData(detail?.requestData)}
        </Typography.Paragraph>
      </Modal>
    </>
  );
};

export default OperLogIndex;