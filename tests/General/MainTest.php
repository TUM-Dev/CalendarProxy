<?php

class MainTest extends \PHPUnit\Framework\TestCase {

    public function testDupeDetection() {
        $data = [
            [
                'DTSTART'  => 'DTSTART:20170627T081500Z',
                'DTEND'    => 'DTSTART:20170627T094500Z',
                'SUMMARY'  => '123test',
                'LOCATION' => '1',
            ],
            [
                'DTSTART'  => 'DTSTART:20170627T081500Z',
                'DTEND'    => 'DTSTART:20170627T094500Z',
                'SUMMARY'  => '123test',
                'LOCATION' => '2',
            ],
            [
                'DTSTART'  => 'DTSTART:20170627T081500Z',
                'DTEND'    => 'DTSTART:20170627T094500Z',
                'SUMMARY'  => '123test',
                'LOCATION' => '3',
            ],
        ];

        CalProxy\handler::noDupes($data);

        $this->assertEquals(1, count($data));
        $this->assertEquals([
            'DTSTART'  => 'DTSTART:20170627T081500Z',
            'DTEND'    => 'DTSTART:20170627T094500Z',
            'SUMMARY'  => '123test',
            'LOCATION' => "3\n2\n1",
        ], array_pop($data));
    }

    public function testNoDupeDetection() {
        $data = [
            [
                'DTSTART'  => 'DTSTART:20170627T071500Z',
                'DTEND'    => 'DTSTART:20170627T094500Z',
                'SUMMARY'  => '123test',
                'LOCATION' => '1',
            ],
            [
                'DTSTART'  => 'DTSTART:20170627T081500Z',
                'DTEND'    => 'DTSTART:20170627T094500Z',
                'SUMMARY'  => '123test',
                'LOCATION' => '2',
            ],
            [
                'DTSTART'  => 'DTSTART:20170627T091500Z',
                'DTEND'    => 'DTSTART:20170627T094500Z',
                'SUMMARY'  => '123test',
                'LOCATION' => '3',
            ],
        ];

        CalProxy\handler::noDupes($data);

        $this->assertEquals(3, count($data));
    }

    public function testNoDupeDetectionOnSummary() {
        $data = [
            [
                'DTSTART'  => 'DTSTART:20170627T081500Z',
                'DTEND'    => 'DTSTART:20170627T094500Z',
                'SUMMARY'  => 'test1',
                'LOCATION' => '1',
            ],
            [
                'DTSTART'  => 'DTSTART:20170627T081500Z',
                'DTEND'    => 'DTSTART:20170627T094500Z',
                'SUMMARY'  => 'test2',
                'LOCATION' => '2',
            ],
            [
                'DTSTART'  => 'DTSTART:20170627T081500Z',
                'DTEND'    => 'DTSTART:20170627T094500Z',
                'SUMMARY'  => 'test3',
                'LOCATION' => '3',
            ],
        ];

        CalProxy\handler::noDupes($data);

        $this->assertEquals(3, count($data));
    }
}