##Syncd-console（syncd命令行插件）
使用步骤:

1.可执行程序当前目录配置 syncd-console.ini 
(可不配置，第一次运行程序按向导提示将自动完成)
```
schema = http
host = <<your syncd host>>
username = <<username>>
password = <<password>>
```
2.登录Syncd
```
./syncd-console login
```

3.查看可发布任务列表(用于部署中的project-name)
```
./syncd-console projects
```

4.查看当前已提交任务列表
```
./syncd-console tasks
```

5.一键部署
```
./syncd-console submit -p <<project-name>> -m "描述"
./syncd-console submit -p <<project-name>> -m "描述" -t "tag"
```

Author: 7853151@qq.com
