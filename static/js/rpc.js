/**
 * Created by ash on 2016/11/4.
 */

function GetValueFromArray(s, k) {
    de = 0.0;
    for (i in s) {
        if (s[i].key == k) {
            return s[i].doc_count
        }
    }
    return de
}

function loadrpclog(app, mychart, option, text, flag, key_map, key_array) {

    $.ajax({
        type: "post",
        async: true,
        url: "/rpc/historydata",
        data: {
            "type": text,
            "sync": flag
        },
        dataType: "json", //类型为数组
        success: function (result) {
            for (i in result) {
                ts_d = result[i].value
                for (k in ts_d) {
                    if (!key_map.has(ts_d[k].key)) {
                        key_map.set(ts_d[k].key, ts_d[k].key)
                    }
                }

            }

            key_map.forEach(function (item, key) {
                key = key.substring(0, 10)
                key_array.push(key)
            })
            var ti = null;
            switch (text + flag) {
                case "rpctrue":
                    ti = "上海发货同步";
                    break;
                case "rpcfalse":
                    ti = "上海发货异步";
                    break;
                case "taobaotrue":
                    ti = "淘宝发货同步";
                    break;
                case "taobaofalse":
                    ti = "淘宝发货异步";
                    break;
            }

            option = {
                title: {
                    text: ti
                },
                legend: {
                    data: key_array
                },
                tooltip: {
                    trigger: 'axis'
                },
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
                xAxis: [{
                    axisLabel: {
                        formatter: function (val) {
                            return val.split(" ").join("\n");
                        },
                    },
                    data: (function () {
                        var xaxis_data = [];
                        for (i in result) {
                            // xaxis_data.unshift(result[i].time.split(" ")[1])
                            xaxis_data.unshift(result[i].time)
                        }
                        return xaxis_data;
                    })()
                }],
                yAxis: {},
                series: (function () {
                    var res = [];
                    for (i in key_array) {
                        res.push({
                            type: "line",
                            name: key_array[i],
                            data: (function () {
                                var series_data = [];
                                for (n in result) {
                                    find = false;
                                    for (k in result[n].value) {
                                        if (result[n].value[k].key == key_array[i]) {
                                            find = true
                                            series_data.unshift(result[n].value[k].doc_count)
                                        }
                                    }

                                    if (!find) {
                                        series_data.unshift(0)
                                    }
                                }
                                return series_data;
                            })()
                        })
                    }
                    return res
                })()
            };

            mychart.setOption(option);
            mychart.hideLoading();

        },
        error: function (errorMsg) {
            //请求失败时执行该函数
            console.log(errorMsg)
            mychart.showLoading()

        }
    })
    app.timeTicket = setInterval(function () {
        load_real_time(mychart, option, text, flag, key_map, key_array)
    }, 5000 * 12);
}


function load_real_time(mychart, option, text, flag, key_map, key_array) {

    $.ajax({
        type: "post",
        async: true,
        url: "/rpc/historydata",
        data: {
            "type": text,
            "sync": flag
        },
        dataType: "json",
        success: function (result) {
            if (result) {
                old_time = option.xAxis[0].data.pop();
                //new_time = result[0].time.split(" ")[1]
                new_time = result[0].time
                if (old_time != new_time) {
                    option.xAxis[0].data.shift();
                    option.xAxis[0].data.push(old_time);
                    option.xAxis[0].data.push(new_time);
                    d = result[0].value

                    for (k in d) {
                        if (!key_map.has(d[k].key)) {
                            key_map.set(d[k].key, d[k].key)
                            key_array.push(d[k].key.substring(0, 10))
                        }
                    }
                    //TODO 少了某类型数据的处理方法
                    for (r in key_array) {
                        find = false;
                        for (var i = 0; i < option.series.length; i++) {
                            gvfa = GetValueFromArray(d, key_array[r])

                            if (option.series[i].name == key_array[r]) {
                                find = true
                                option.series[i].data.pop();
                                option.series[i].data.push(gvfa);

                            }
                        }

                        //增加一个新的线条
                        if (!find) {
                            option.legend.data.push(key_array[r])
                            newob = {
                                type: "line",
                                name: rel_key_array[r],
                                data: [GetValueFromArray(d, key_array[r])]
                            }
                            option.push(newob)
                        }
                    }
                    mychart.setOption(option);
                } else {
                    option.xAxis[0].data.push(old_time)
                }
                //分组
            }

        },
        error: function (errorMsg) {
            //请求失败时执行该函数
            alert("历史数据处理失败!");

        }
    })
}