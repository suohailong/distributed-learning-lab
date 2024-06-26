# 分布式锁设计

## 功能特性
1. 锁必须是可重入锁
2. 锁必须是公平锁

## 非功能特性
1. 锁必须可靠， 避免单点故障


## 详细设计
1. 公平锁设计-设计1 
   1. 加锁
    ```mermaid
    flowchart TB
        a0[开始] --> a1[setnx lockkey: 加锁]
        a1 --> a2{返回值是否为1}
        a2 --1--> a3[结束]
        a2 --0--> a4[rpush lockkey_wait: 加入等待队列]
        a4 --> a5[subscribe lockkey_channel: 等待解锁通知]
        a5 --> a6[收到订阅消息]
        a6 --> a7{判断锁信息是否为自己}
        a7 --是--> a1
        a7 --不是-->a5

    ```

    2. 解锁
    ```mermaid
    flowchart TB
        a6[开始] --> a7[del lockkey: 解锁]
        a7 --> a8[rpop lockkey_wait: 取出等待的锁]
        a8 --> a11{rpop是否为空}
        a11 --NO--> a9[publish lockkey_channel 锁信息: 通知锁释放]
        a9 --> a10[结束]
        a11 --Yes--> a10

    ```
    由于加锁和解锁设计多个redis 操作，为了解决操作的原子性， 我们用lua脚本实现以上加锁和解锁的逻辑

3. 重入锁设计
   1. 加锁
   ```mermaid
   flowchart TB
        a1[开始] --> a2{本地持有且值相同}
        a2 --YES--> a9[本地计数器+1]
        a9 --> a3[加锁成功]
        a2 --NO--> a5[redis加锁]
        a5 --> a6{成功?}
        a6 --YES--> a7[本地计数器+1]
        a7 --> a10[加锁成功]
        a10 --> a8[结束]
        a5 --> a11[重试]
        a11 --超过重试次数--> a8
        a3 --> a8[结束]
   ```
   2. 解锁
   ```mermaid
   flowchart TB
        a1[开始] --> a2{本地持有值相同且计数>0}
        a2 --YES--> a9[本地计数器-1]
        a9 --> a3[解锁成功]
        a2 --NO--> a5[redis解锁]
        a5 --> a6{成功?}
        a6 --YES--> a7[本地计数器-1]
        a7 --> a10[解锁成功]
        a10 --> a8[结束]
        a5 --> a11[重试]
        a11 --超过重试次数--> a8
        a3 --> a8[结束]
   ```
4. 可靠性设计 
    可靠性的设计要解决的问题是redis单点故障，解决思路也很简单就是在以上设计基础上在多个redis实例上申请相同的锁。当大多数redis加锁成功则认为客户端获取锁成功。 具体流程如下：
    1. 客户端A记录当前申请时间T1，并向全部redis 实例申请加锁
    2. redis实例返回结果， 客户端判断是否大多数申请成功
    3. 如果大多数返回成功， 记录当前时间戳T2， 判断T2-T1是否小于锁的超时时间
    4. 小于则加锁成功， 否者加锁失败。
    5. 加锁失败，向全部节点发送释放锁请求 

    说明：
        第3步， 判断T2-T1是否小于锁的超时时间， 是为了防止NPC问题中的N和P,即长时间的网络延迟或者程序暂停。 如果不判断，这两个问题会导致虽然加锁成功但实际上锁已经过期
        第5步， 必须向全部节点发送释放锁的请求， 因为有些返回失败的节点可能是因为网络问题导致响应失败，其实锁已经申请成功， 如果不去释放会导致其他来此节点加锁的客户端加锁不成功

    缺点： 
        1. redis实例时钟必须一致, 否者就会出现锁在各个实例的过期时间点不一致。 举例如下:
            假设a 通过实例A, B,C获取了锁， 但由于网络问题无法访问D和E
            节点C 上的时钟发生跳跃导致C 上的锁过期
            b 由于网络问题无法访问 A和B， b 通过C D E获取到了锁， 此时 双方都认为自己有锁
        2. 无法解决加锁成功后由于网络中断或者长时间暂停导致的锁失效却仍然在处理共享资源的问题

5. 续约 
    续约是为了防止当客户端在长时间业务处理的情况锁被超期释放。 通过续约可以保证客户端在处理完业务的情况下才会释放锁
    
