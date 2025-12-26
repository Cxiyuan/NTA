import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Statistic,
  Table,
  Tag,
  Space,
  Button,
  Progress,
  Alert,
  Spin,
  Tooltip,
} from 'antd';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  SyncOutlined,
  ThunderboltOutlined,
  CloudServerOutlined,
  DatabaseOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { Line } from '@ant-design/plots';
import apiClient from '../utils/apiClient';

interface KafkaTopic {
  name: string;
  partitions: number;
  messages: number;
  lag: number;
}

interface ConsumerGroup {
  group_id: string;
  members: number;
  lag: number;
  state: string;
}

interface FlinkJob {
  job_id: string;
  name: string;
  status: string;
  start_time: string;
  duration: string;
  tasks: {
    total: number;
    running: number;
    failed: number;
  };
}

const StreamProcessing: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [kafkaTopics, setKafkaTopics] = useState<KafkaTopic[]>([]);
  const [consumerGroups, setConsumerGroups] = useState<ConsumerGroup[]>([]);
  const [flinkJobs, setFlinkJobs] = useState<FlinkJob[]>([]);
  const [kafkaHealth, setKafkaHealth] = useState<string>('healthy');
  const [flinkHealth, setFlinkHealth] = useState<string>('healthy');
  const [throughputData, setThroughputData] = useState<any[]>([]);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [topicsRes, groupsRes, jobsRes] = await Promise.all([
        apiClient.get('/api/v1/stream/kafka/topics'),
        apiClient.get('/api/v1/stream/kafka/consumer-groups'),
        apiClient.get('/api/v1/stream/flink/jobs'),
      ]);

      setKafkaTopics(topicsRes.data.topics || []);
      setConsumerGroups(groupsRes.data.groups || []);
      setFlinkJobs(jobsRes.data.jobs || []);
      
      // Simulate throughput data
      const now = Date.now();
      setThroughputData(prev => {
        const newData = [...prev, {
          time: new Date(now).toLocaleTimeString(),
          messages: Math.floor(Math.random() * 10000) + 5000,
        }].slice(-20);
        return newData;
      });
    } catch (error) {
      console.error('Failed to fetch stream data:', error);
    } finally {
      setLoading(false);
    }
  };

  const topicColumns = [
    {
      title: 'Topic名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => <Tag color="blue">{text}</Tag>,
    },
    {
      title: '分区数',
      dataIndex: 'partitions',
      key: 'partitions',
    },
    {
      title: '消息数',
      dataIndex: 'messages',
      key: 'messages',
      render: (val: number) => val.toLocaleString(),
    },
    {
      title: '积压',
      dataIndex: 'lag',
      key: 'lag',
      render: (lag: number) => (
        <span style={{ color: lag > 100 ? '#ff4d4f' : '#52c41a' }}>
          {lag.toLocaleString()}
        </span>
      ),
    },
  ];

  const consumerColumns = [
    {
      title: '消费组',
      dataIndex: 'group_id',
      key: 'group_id',
    },
    {
      title: '成员数',
      dataIndex: 'members',
      key: 'members',
    },
    {
      title: '总积压',
      dataIndex: 'lag',
      key: 'lag',
      render: (lag: number) => (
        <Tag color={lag === 0 ? 'green' : lag < 100 ? 'orange' : 'red'}>
          {lag}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'state',
      key: 'state',
      render: (state: string) => (
        <Tag color={state === 'Stable' ? 'green' : 'red'}>{state}</Tag>
      ),
    },
  ];

  const jobColumns = [
    {
      title: '作业名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const color = status === 'RUNNING' ? 'green' : status === 'FAILED' ? 'red' : 'orange';
        const icon = status === 'RUNNING' ? <SyncOutlined spin /> : <CloseCircleOutlined />;
        return (
          <Tag color={color} icon={icon}>
            {status}
          </Tag>
        );
      },
    },
    {
      title: '运行时长',
      dataIndex: 'duration',
      key: 'duration',
    },
    {
      title: 'Task进度',
      key: 'progress',
      render: (record: FlinkJob) => {
        const percent = (record.tasks.running / record.tasks.total) * 100;
        return (
          <Progress
            percent={percent}
            size="small"
            status={record.tasks.failed > 0 ? 'exception' : 'active'}
            format={() => `${record.tasks.running}/${record.tasks.total}`}
          />
        );
      },
    },
    {
      title: '操作',
      key: 'action',
      render: (record: FlinkJob) => (
        <Space>
          <Button size="small" type="link">详情</Button>
          <Button size="small" type="link" danger>取消</Button>
        </Space>
      ),
    },
  ];

  const throughputConfig = {
    data: throughputData,
    xField: 'time',
    yField: 'messages',
    smooth: true,
    animation: {
      appear: {
        animation: 'path-in',
        duration: 1000,
      },
    },
    color: '#1890ff',
    lineStyle: {
      lineWidth: 2,
    },
  };

  const totalMessages = kafkaTopics.reduce((sum, topic) => sum + topic.messages, 0);
  const totalLag = kafkaTopics.reduce((sum, topic) => sum + topic.lag, 0);

  return (
    <div style={{ padding: 24 }}>
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <h2 style={{ marginBottom: 16 }}>
            <ThunderboltOutlined /> 流处理监控
            <Button
              icon={<ReloadOutlined />}
              style={{ float: 'right' }}
              onClick={fetchData}
              loading={loading}
            >
              刷新
            </Button>
          </h2>
        </Col>

        {/* 健康状态卡片 */}
        <Col span={12}>
          <Card>
            <Statistic
              title="Kafka集群状态"
              value={kafkaHealth === 'healthy' ? '正常' : '异常'}
              prefix={
                kafkaHealth === 'healthy' ? (
                  <CheckCircleOutlined style={{ color: '#52c41a' }} />
                ) : (
                  <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
                )
              }
              valueStyle={{ color: kafkaHealth === 'healthy' ? '#52c41a' : '#ff4d4f' }}
            />
          </Card>
        </Col>

        <Col span={12}>
          <Card>
            <Statistic
              title="Flink集群状态"
              value={flinkHealth === 'healthy' ? '正常' : '异常'}
              prefix={
                flinkHealth === 'healthy' ? (
                  <CheckCircleOutlined style={{ color: '#52c41a' }} />
                ) : (
                  <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
                )
              }
              valueStyle={{ color: flinkHealth === 'healthy' ? '#52c41a' : '#ff4d4f' }}
            />
          </Card>
        </Col>

        {/* Kafka统计 */}
        <Col span={6}>
          <Card>
            <Statistic
              title="Topic总数"
              value={kafkaTopics.length}
              prefix={<DatabaseOutlined />}
            />
          </Card>
        </Col>

        <Col span={6}>
          <Card>
            <Statistic
              title="消息总数"
              value={totalMessages}
              formatter={(value) => value.toLocaleString()}
            />
          </Card>
        </Col>

        <Col span={6}>
          <Card>
            <Statistic
              title="消费积压"
              value={totalLag}
              valueStyle={{ color: totalLag > 1000 ? '#ff4d4f' : '#52c41a' }}
            />
          </Card>
        </Col>

        <Col span={6}>
          <Card>
            <Statistic
              title="Flink作业数"
              value={flinkJobs.length}
              prefix={<CloudServerOutlined />}
            />
          </Card>
        </Col>

        {/* 吞吐量图表 */}
        <Col span={24}>
          <Card title="消息吞吐量 (消息/秒)">
            <Line {...throughputConfig} height={200} />
          </Card>
        </Col>

        {/* Kafka Topics */}
        <Col span={24}>
          <Card title="Kafka Topics" extra={<Tag color="blue">{kafkaTopics.length} 个Topic</Tag>}>
            <Table
              columns={topicColumns}
              dataSource={kafkaTopics}
              rowKey="name"
              pagination={false}
              size="small"
            />
          </Card>
        </Col>

        {/* Consumer Groups */}
        <Col span={24}>
          <Card title="消费者组" extra={<Tag color="green">{consumerGroups.length} 个消费组</Tag>}>
            <Table
              columns={consumerColumns}
              dataSource={consumerGroups}
              rowKey="group_id"
              pagination={false}
              size="small"
            />
          </Card>
        </Col>

        {/* Flink Jobs */}
        <Col span={24}>
          <Card
            title="Flink作业"
            extra={
              <Space>
                <Tag color="green">{flinkJobs.filter(j => j.status === 'RUNNING').length} Running</Tag>
                <Tag color="red">{flinkJobs.filter(j => j.status === 'FAILED').length} Failed</Tag>
              </Space>
            }
          >
            <Table
              columns={jobColumns}
              dataSource={flinkJobs}
              rowKey="job_id"
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default StreamProcessing;
