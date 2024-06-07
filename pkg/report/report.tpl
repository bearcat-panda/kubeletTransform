<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>服务器预检报告</title>
	<style>
		body {
			font-family: Arial, sans-serif;
			font-size: 14px;
			color: #333;
			margin: 0;
			padding: 0;
		}
		
		h1 {
			font-size: 28px;
			font-weight: bold;
			margin: 20px 0;
			text-align: center;
			color: #333;
		}
		
		h2 {
			font-size: 24px;
			font-weight: bold;
			margin: 20px 0 10px;
			color: #333;
		}

        .flex-container {
  			display: flex;
			font-size: 20px;
			justify-content: space-between;
			align-items: center;
			margin-bottom: 2px;
		}
		
		table {
			border-collapse: collapse;
			width: 100%;
			margin-bottom: 20px;
			box-shadow: 0 0 20px rgba(0, 0, 0, 0.1);
		}
		
		th, td {
			text-align: left;
			padding: 8px;
			border: 1px solid #ddd;
			font-size: 14px;
			color: #333;
		}
		
		tr:nth-child(even) {
			background-color: #f2f2f2;
		}
		
		th {
			background-color: #4c7aaf;
			color: white;
			font-weight: bold;
		}
		
		.footer {
			font-size: 12px;
			color: #999;
			text-align: center;
			margin-top: 20px;
		}
	</style>
</head>
<body>
	<h1>检查报告</h1>	
	<h2>统计信息</h2>
	<div class="flex-container">
	<div> 开始时间：{{.StartTime}} </div>
	<div> 持续时间：{{.DurationTime}} </div>
	<div> 随机种子：{{.RandomSeed}} </div>
	</div>
	<table>
		<tr>
			<th>总测试数</th>
			<th>成功数</th>
			<th>失败数</th>
			<th>警告数</th>
			<th>测试结果</th>
		</tr>
		<tr>
			<td>{{.Total}}</td>
			<td>{{.Success}}</td>
			<td>{{.Failure}}</td>
			<td>{{.Warning}}</td>
            {{if eq .Result "pass"}}
			<td style="color:rgb(61, 47, 255)">通过</td>
            {{else}}
            <td style="color:red">不通过</td>
            {{end}}
		</tr>
	</table>
	
	<h2>测试结果</h2>
	<table>
		<tr>
		    <th>节点</th>
		    <th>角色</th>
			<th>检查任务项</th>
			<th>检查结果</th>
			<th>耗时</th>
			<th>详细信息</th>
		</tr>
        {{range .Case}}
        <tr>
            <td>{{.IP}}</td>
            <td>{{.Role}}</td>
			<td>{{.Name}}</td>
            {{if eq .Status "Success"}}
		    <td style="color:green">成功</td>
            {{else if eq .Status "Warning"}}
            <td style="color:orange">警告</td>
            {{else}}
            <td style="color:red">失败</td>
            {{end}}
            <td>{{.DurationTime}}</td>
			<td>{{.Detail}}</td>
		</tr>
        {{end}}
	</table>
	
	<h2>服务器信息</h2>
	<table>
		<tr>
		    <th>节点</th>
		    <th>角色</th>
			<th>操作系统</th>
			<th>内核</th>
			<th>CPU</th>
			<th>内存</th>
			<th>网卡</th>
			<th>磁盘</th>
		</tr>
        {{range .Server}}
        <tr>
            <td>{{.IP}}</td>
            <td>{{.Role}}</td>
			<td>{{.OS}}</td>
			<td>{{.Kernel}}</td>
			<td>{{.CPU}}</td>
			<td>{{.Memory}}</td>
            <td>{{.Network}}</td>
			<td>{{.Disk}}</td>
		</tr>
        {{end}}
	</table>
</body>
</html>
