package main

  import (
    "fmt"
  	"github.com/aws/aws-sdk-go/aws"
  	"github.com/aws/aws-sdk-go/aws/awserr"
  	"github.com/aws/aws-sdk-go/aws/session"
  	"github.com/aws/aws-sdk-go/aws/endpoints"
  	"github.com/aws/aws-sdk-go/service/ec2"
  )

func delete_eni(svc *ec2.EC2, eni string) bool {
    input := &ec2.DeleteNetworkInterfaceInput{
        NetworkInterfaceId: aws.String(eni),
    }

    _, err := svc.DeleteNetworkInterface(input)
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok {
            switch aerr.Code() {
            default:
                fmt.Println(aerr.Error())
            }
        } else {
            fmt.Println(err.Error())
        }
        return false
    }
	return true
}

func create_parking_eni(svc *ec2.EC2, subnet string, sg string) string {
    input := &ec2.CreateNetworkInterfaceInput{
        Description: aws.String("my network interface"),
        Groups: []*string{
            aws.String(sg),
        },
        SubnetId:         aws.String(subnet),
    }

    result, err := svc.CreateNetworkInterface(input)
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok {
            switch aerr.Code() {
            default:
                fmt.Println(aerr.Error())
            }
        } else {
            fmt.Println(err.Error())
        }
        return ""
    }
	return *result.NetworkInterface.NetworkInterfaceId
}

func get_secips_from_eni(svc *ec2.EC2, eni string) map[string]bool {
	set := make(map[string]bool)
    input := &ec2.DescribeNetworkInterfacesInput{
        NetworkInterfaceIds: []*string {
		aws.String(eni),
		},
    }
    result, err := svc.DescribeNetworkInterfaces(input)
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok {
            switch aerr.Code() {
            default:
                fmt.Println(aerr.Error())
            }
        } else {
            fmt.Println(err.Error())
        }
        return set
    }
	if len(result.NetworkInterfaces) > 0 {
		pips := result.NetworkInterfaces[0].PrivateIpAddresses
		for i := 0; i < len(pips); i++ {
			if ! *pips[i].Primary {
				set[*pips[i].PrivateIpAddress] = true
			}
		}
	}
	return set
}

func allocate_private_vip(svc *ec2.EC2, eni string) string {
	old := get_secips_from_eni(svc, eni)
	input := &ec2.AssignPrivateIpAddressesInput{
    	NetworkInterfaceId:             aws.String(eni),
    	SecondaryPrivateIpAddressCount: aws.Int64(1),
	}
	_, err := svc.AssignPrivateIpAddresses(input)
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok {
            switch aerr.Code() {
            default:
                fmt.Println(aerr.Error())
            }
        } else {
            fmt.Println(err.Error())
        }
        return ""
    }
	new := get_secips_from_eni(svc, eni)
	for k, _ := range new {
		if _, ok := old[k]; !ok {
			// TODO VS IPs checks needs to be added
			return k
		}
	}
	return ""
}

func move_ip_to_eni(svc *ec2.EC2, eni string, ip string) bool {
    input := &ec2.AssignPrivateIpAddressesInput{
        NetworkInterfaceId: aws.String(eni),
        PrivateIpAddresses: []*string{
            aws.String(ip),
        },
		AllowReassignment: aws.Bool(true),
    }

    _, err := svc.AssignPrivateIpAddresses(input)
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok {
            switch aerr.Code() {
            default:
                fmt.Println(aerr.Error())
            }
        } else {
            fmt.Println(err.Error())
        }
        return false
    }
	return true
}

func remove_ip_from_eni(svc *ec2.EC2, eni string, ip string) bool {
    input := &ec2.UnassignPrivateIpAddressesInput{
        NetworkInterfaceId: aws.String(eni),
        PrivateIpAddresses: []*string{
            aws.String(ip),
        },
    }

    _, err := svc.UnassignPrivateIpAddresses(input)
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok {
            switch aerr.Code() {
            default:
                fmt.Println(aerr.Error())
            }
        } else {
            fmt.Println(err.Error())
        }
        return false
    }
	return true
}

func get_sgid_from_vm(svc *ec2.EC2, vm_id string) string {
	input := &ec2.DescribeInstanceAttributeInput{
		Attribute:  aws.String("groupSet"),
		InstanceId: aws.String(vm_id),
	}

	result, err := svc.DescribeInstanceAttribute(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return ""
	}
	if len(result.Groups) > 0 {
		return *result.Groups[0].GroupId
	}
	return ""
}

func get_sg_from_sgid(svc *ec2.EC2, sgid string) {
    input := &ec2.DescribeSecurityGroupsInput{
        GroupIds: []*string {
		aws.String(sgid),
		},
    }
    result, err := svc.DescribeSecurityGroups(input)
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok {
            switch aerr.Code() {
            default:
                fmt.Println(aerr.Error())
            }
        } else {
            fmt.Println(err.Error())
        }
        return
    }
	fmt.Println(result)
}

func print_error(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		default:
			fmt.Println(aerr.Error())
		}
	} else {
		fmt.Println(err.Error())
	}
}

func create_sg(svc *ec2.EC2, name string) string {
	input := &ec2.CreateSecurityGroupInput{
		Description: aws.String("test-creation"),
		GroupName: aws.String(name),
	}
	result, err := svc.CreateSecurityGroup(input)
	if err == nil {
		return *result.GroupId
	} else {
		print_error(err)
	}
	return ""
}

func goroutine(svc *ec2.EC2) {
	//subnet  := "subnet-62f1b707"
	//sg := "sg-a9a00fcd"
	fmt.Println("Go routine called")
	eni := "eni-00bce3ae76a4f9bf4"
	//eni2 := create_parking_eni(svc, subnet, sg)
	eni2 := "eni-02218a33a8fb51aba"
	vip := allocate_private_vip(svc, eni)
	if vip != "" {
		fmt.Println(vip)
		fmt.Println(eni2)
		if move_ip_to_eni(svc, eni2, vip) {
			fmt.Println("Moving IP succeded")
			remove_ip_from_eni(svc, eni2, vip)
		}
	}
}

func callgoroutine(svc *ec2.EC2) {
	for i := 50000; i < 50010; i++ {
		go gor(svc, int64(i))
	}
	fmt.Scanln()
}

func gor(svc *ec2.EC2, port int64) {
	update_sg(svc, "sg-0020e426da1e56d8d", port)
	get_sg_from_sgid(svc, "sg-0020e426da1e56d8d")
}

func update_sg(svc *ec2.EC2, sgid string, port int64) {
	input := &ec2.AuthorizeSecurityGroupIngressInput{
		CidrIp: aws.String("0.0.0.0/0"),
		FromPort: aws.Int64(port),
		IpProtocol: aws.String("tcp"),
		ToPort: aws.Int64(port),
		GroupId: aws.String(sgid),
	}
	result, err := svc.AuthorizeSecurityGroupIngress(input)
	if err == nil {
		fmt.Println(result)
	} else {
		print_error(err)
	}
}

func main() {
  	sess := session.Must(session.NewSession(&aws.Config{
      Region: aws.String(endpoints.UsWest2RegionID),}))
  	svc := ec2.New(sess)
	callgoroutine(svc)
}
