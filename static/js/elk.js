/**
 * Created by ash on 2016/11/4.
 */

function loadoption(app, myChart, text, option) {
    $.ajax({
        type: "post",
        async: true, //异步请求（同步请求将会锁住浏览器，用户其他操作必须等待请求完成才可以执行）
        url: "/elk_monitor/historydata", //请求发送到TestServlet处
        data: {
            "type": text
        },
        dataType: "json", //类型为数组
        success: function (result) {
            if (result) {
                option = {
                    title: {
                        text: text,
                        subtext: '5分钟刷新'
                    },
                    tooltip: {
                        trigger: 'axis'
                    },
                    legend: {
                        data: ["总数", "失败数"]
                    },
                    xAxis: [{
                        // type: 'category',
                        // boundaryGap: true,
                        axisLabel: {
                            formatter: function (val) {
                                return val.split(" ").join("\n");
                            },
                        },
                        data: (function () {
                            var res = [];
                            for (i in result) {
                                res.push(result[i].time)
                            }
                            return res;
                        })()
                    }],
                    yAxis: [{
                        type: 'value',
                        name: '总数',
                        max: 'dataMax'
                    }],
                    toolbox: {
                        show: true,
                        feature: {
                            mark: {
                                show: true
                            },
                            dataView: {
                                show: true,
                                readOnly: false
                            },
                            magicType: {
                                show: true,
                                type: ['bar', 'stack', 'tiled']
                            },
                            restore: {
                                show: true
                            },
                            saveAsImage: {
                                show: true
                            }
                        }
                    },
                    calculable: true,
                    series: [{
                        type: "line",
                        name: "总数",
                        itemStyle: {
                            normal: {
                                lineStyle: {
                                    color: '#AE0000'
                                }
                            }
                        },
                        // itemStyle: {normal: {areaStyle: {type: 'default'}}},
                        smooth: true,
                        data: (function () {
                            var res = [];
                            for (i in result) {
                                res.push(result[i].value[1])
                            }
                            return res;
                        })(),
                        markPoint: {
                            data: [{
                                type: 'max',
                                name: '最大值'
                            }, {
                                type: 'min',
                                name: '最小值'
                            }]
                        },
                    }, {
                        type: "line",
                        name: "失败数",
                        smooth: true,
                        itemStyle: {
                            normal: {
                                lineStyle: {
                                    color: '#6C6C6C'
                                }
                            }
                        },
                        //itemStyle: {normal: {areaStyle: {type: 'default'}}},//堆面积
                        data: (function () {
                            var res = [];
                            for (i in result) {
                                res.push(result[i].value[0])
                            }
                            return res;
                        })(),
                        markPoint: {
                            data: [{
                                type: 'max',
                                name: '最大值'
                            }, {
                                type: 'min',
                                name: '最小值'
                            }]
                        },
                    }]
                };
                myChart.setOption(option)
            }
        },
        error: function (errorMsg) {
            //请求失败时执行该函数
            alert("历史数据处理失败!");
        }
    });
    app.timeTicket = setInterval(function () {
        realtimeajax(myChart, option, text)
    }, 5000 * 12);;

    //取实时数据
}


function realtimeajax(myChart, option, text) {
    $.ajax({
        type: "post",
        async: true,
        url: "/elk_monitor/historydata",
        data: {
            "type": text
        },
        dataType: "json", //返回数据形式为json
        success: function (result) {
            if (result) {
                old_time = option.xAxis[0].data.pop();
                if (old_time != result[11].time) {
                    option.xAxis[0].data.push(old_time) //将较新的时间写入
                    option.xAxis[0].data.shift() //弹出最老的时间
                    option.xAxis[0].data.push(result[11].time); //写入最新的时间
                    var data0 = option.series[0].data;
                    var data1 = option.series[1].data;
                    data1.shift();
                    data1.push(result[11].value[0]);
                    data0.shift();
                    data0.push(result[11].value[1]);
                    myChart.setOption(option);
                } else {
                    option.xAxis[0].data.push(old_time)
                }
            }
        },
        error: function (errorMsg) {
            //请求失败时执行该函数
            alert("图表请求数据失败!");
        }
    })
}