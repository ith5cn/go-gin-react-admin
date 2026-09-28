// 本文件由代码生成器生成，重新生成会覆盖手工修改。
import { forwardRef, useImperativeHandle, useState } from "react";
import { Form, Input, InputNumber, Modal, Select, message } from "antd";
import CityLinkage from "@/components/city-linkage";
import { aiArticleCreateApi, aiArticleUpdateApi } from "@/api/system/ai-article";

export interface AiarticleEditRef {
  open: (type?: "add" | "edit", data?: Record<string, unknown>) => void;
}

interface AiarticleEditProps {
  onSuccess?: () => void;
}

const initialFormData = {
  categoryId: [],
  title: "",
  author: "",
  image: "",
  describe: "",
  content: "",
  views: undefined,
  sort: undefined,
  status: undefined,
  isLink: undefined,
  linkUrl: "",
  isHot: undefined,
};

const ARRAY_FIELDS: Array<{ name: string; numeric: boolean }> = [
  { name: "categoryId", numeric: true },
];

const toFormValues = (data: Record<string, unknown>) => {
  const next: Record<string, unknown> = { ...data };
  ARRAY_FIELDS.forEach(({ name, numeric }) => {
    if (typeof next[name] === "string") {
      const items = next[name] === "" ? [] : String(next[name]).split(",");
      next[name] = numeric ? items.map(Number) : items;
    }
  });
  return next;
};

const toSubmitValues = (values: Record<string, unknown>) => {
  const next: Record<string, unknown> = { ...values };
  ARRAY_FIELDS.forEach(({ name }) => {
    if (Array.isArray(next[name])) {
      next[name] = (next[name] as Array<string | number>).join(",");
    }
  });
  return next;
};

const AiarticleEdit = forwardRef<AiarticleEditRef, AiarticleEditProps>(({ onSuccess }, ref) => {
  const [visible, setVisible] = useState(false);
  const [mode, setMode] = useState<"add" | "edit">("add");
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const title = "测试文章菜单" + (mode === "edit" ? " - 编辑" : " - 新增");

  const open = (type: "add" | "edit" = "add", data?: Record<string, unknown>) => {
    setMode(type);
    form.resetFields();
    form.setFieldsValue(type === "edit" && data ? toFormValues(data) : initialFormData);
    setVisible(true);
  };

  const close = () => setVisible(false);

  const handleSubmit = async () => {
    try {
      setLoading(true);
      const values = await form.validateFields();
      const payload = toSubmitValues(values);
      if (mode === "add") {
        await aiArticleCreateApi(payload);
      } else {
        await aiArticleUpdateApi(values.id as number, payload);
      }
      message.success("操作成功");
      onSuccess?.();
      close();
    } catch (error) {
      if ((error as { errorFields?: unknown })?.errorFields) return;
    } finally {
      setLoading(false);
    }
  };

  useImperativeHandle(ref, () => ({ open }));

  return (
    <Modal open={visible} title={title} width={600} confirmLoading={loading} onOk={handleSubmit} onCancel={close}>
      <Form form={form} labelCol={{ span: 4 }} wrapperCol={{ span: 18 }}>
        <Form.Item name="id" hidden>
          <Input />
        </Form.Item>
        <Form.Item name="categoryId" label="分类" rules={[{ required: true, message: "请选择分类" }]}>
          <CityLinkage />
        </Form.Item>
        <Form.Item name="title" label="文章标题" rules={[{ required: true, message: "请输入文章标题" }]}>
          <Input placeholder="请输入文章标题" />
        </Form.Item>
        <Form.Item name="author" label="文章作者">
          <Input placeholder="请输入文章作者" />
        </Form.Item>
        <Form.Item name="image" label="文章图片">
          <Input placeholder="请输入文章图片" />
        </Form.Item>
        <Form.Item name="describe" label="文章简介">
          <Input placeholder="请输入文章简介" />
        </Form.Item>
        <Form.Item name="content" label="文章内容">
          <Input.TextArea rows={4} placeholder="请输入文章内容" />
        </Form.Item>
        <Form.Item name="views" label="浏览次数">
          <InputNumber style={{ width: "100%" }} placeholder="请输入浏览次数" />
        </Form.Item>
        <Form.Item name="sort" label="排序">
          <InputNumber style={{ width: "100%" }} placeholder="请输入排序" />
        </Form.Item>
        <Form.Item name="status" label="状态">
          <Select allowClear options={[{"label":"正常","value":1},{"label":"停用","value":2}]} />
        </Form.Item>
        <Form.Item name="isLink" label="是否外链">
          <Select allowClear options={[{"label":"是","value":1},{"label":"否","value":2}]} />
        </Form.Item>
        <Form.Item name="linkUrl" label="链接地址">
          <Input placeholder="请输入链接地址" />
        </Form.Item>
        <Form.Item name="isHot" label="是否热门">
          <Select allowClear options={[{"label":"是","value":1},{"label":"否","value":2}]} />
        </Form.Item>
      </Form>
    </Modal>
  );
});

AiarticleEdit.displayName = "AiarticleEdit";

export default AiarticleEdit;
